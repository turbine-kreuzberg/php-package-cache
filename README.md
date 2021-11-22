# php-package-cache

Local composer cache intended as a pull though cache to speed up php builds in ci pipelines.

## Usage examples

Use repman plugin to rewrite requests
```
composer global require repman-io/composer-plugin
```

Add config to `composer.json`, to fetch files via proxy
```
{
    "extra": {
        "repman": {
            "url": "http://php-package-cache:8080/"
        }
    }
}
```

Optional add config to `composer.json`, to fetch internally via http
```
{
    "config": {
        "secure-http": false
    }
}
```
