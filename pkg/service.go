package ppc

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	log_ "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/turbine-kreuzberg/php-package-cache/pkg/adapter"
)

// ServiceServer executes the users requests.
type ServiceServer struct {
	router  *mux.Router
	storage *adapter.ObjectStorage
}

// NewServiceServer creates a new service server and initiates the routes.
func NewService(storage *adapter.ObjectStorage) http.Handler {
	router := mux.NewRouter()
	srv := &ServiceServer{
		router:  router,
		storage: storage,
	}
	srv.addRoutes()
	return srv
}

func (s *ServiceServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *ServiceServer) addRoutes() {
	s.router.HandleFunc("/healthz/ready", Ok)
	s.router.HandleFunc("/healthz/alive", Ok)
	s.router.Handle("/metrics", promhttp.Handler())

	s.router.HandleFunc("/downloads", Downloads)

	s.router.HandleFunc("/dists/{package_group}/{package_name}/{version_normalized}/{reference}.{type}", s.Package)
}

func Ok(w http.ResponseWriter, r *http.Request) {
	span, _ := opentracing.StartSpanFromContext(r.Context(), "ok")
	defer span.Finish()

	w.WriteHeader(http.StatusOK)
	// only here to return a status 200 http codes
}

func Downloads(w http.ResponseWriter, r *http.Request) {
	span, _ := opentracing.StartSpanFromContext(r.Context(), "report downlads")
	defer span.Finish()

	w.WriteHeader(http.StatusOK)
	// ignored, informs about downloaded packages
	// see https://getcomposer.org/doc/05-repositories.md#notify-batch
}

func (s *ServiceServer) Package(w http.ResponseWriter, r *http.Request) {
	span, ctx := opentracing.StartSpanFromContext(r.Context(), "fetch package")
	defer span.Finish()

	key := r.URL.Path
	vars := mux.Vars(r)
	package_ := fmt.Sprintf("%s/%s", vars["package_group"], vars["package_name"])
	version_normalized := vars["version_normalized"]
	reference := vars["reference"]
	type_ := vars["type"]
	version_normalized_split := strings.Split(version_normalized, ".")
	version_denormalized := fmt.Sprintf("%s.%s.%s", version_normalized_split[0], version_normalized_split[1], version_normalized_split[2])

	span.LogFields(
		log_.String("path", key),
		log_.String("package", package_),
		log_.String("version_normalized", version_normalized),
		log_.String("version_denormalized", version_denormalized),
		log_.String("reference", reference),
		log_.String("type", type_),
	)

	presingedURL, err := s.storage.Presign(ctx, key)
	if err != nil {
		log.Printf("presigning %s: %v", key, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// file is cache
	exists := s.storage.Exists(ctx, key) == nil
	if exists {
		http.Redirect(w, r, presingedURL, http.StatusTemporaryRedirect)
		return
	}

	// url from meta data (packagist)
	url, err := url(ctx, package_, version_normalized, reference, type_)
	if err != nil {
		log.Printf("lookup url for %s: %v", r.URL.Path, err)
	}

	err = cache(ctx, s.storage, key, url)
	if err == nil {
		http.Redirect(w, r, presingedURL, http.StatusTemporaryRedirect)
		return
	}

	log.Printf("caching %s for %s: %v", url, key, err)

	// fallback to repoman
	repomanURL := fmt.Sprintf("https://repo.repoman.io/%s", key)

	err = cache(ctx, s.storage, key, repomanURL)
	if err == nil {
		http.Redirect(w, r, presingedURL, http.StatusTemporaryRedirect)
		return
	}

	log.Printf("caching %s for %s: %v", repomanURL, key, err)

	// fallback to github
	githubURL := fmt.Sprintf("https://codeload.github.com/%s/%s/refs/tags/%v", package_, type_, version_denormalized)

	err = cache(ctx, s.storage, key, githubURL)
	if err == nil {
		http.Redirect(w, r, presingedURL, http.StatusTemporaryRedirect)
		return
	}

	log.Printf("caching %s for %s: %v", githubURL, key, err)

	// send found url from metadata
	if url != "" {
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		return
	}

	// failure
	w.WriteHeader(http.StatusInternalServerError)
}

func cache(ctx context.Context, o *adapter.ObjectStorage, key, url string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cache file")
	defer span.Finish()

	// download
	file, err := fetchFile(ctx, url)
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	// upload
	err = o.Upload(ctx, key, file, mime.TypeByExtension("zip"))
	if err != nil {
		return err
	}

	return nil
}

// Caller is required to cleanup the returned file.
func fetchFile(ctx context.Context, url string) (*os.File, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "download file")
	defer span.Finish()

	client := httpClient()

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("http status code %d", resp.StatusCode)
	}

	// create tmp file to convert ReadCloser to ReadSeeker
	tmp, err := ioutil.TempFile("", "php_cache")
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(tmp, resp.Body)
	if err != nil {
		return nil, err
	}
	tmp.Seek(0, os.SEEK_SET)

	return tmp, nil
}
