
disable_snapshots()
allow_k8s_contexts(os.getenv("TILT_ALLOW_CONTEXT"))

include('./dependencies/Tiltfile')
include('./test/Tiltfile')

k8s_yaml('deployment/kubernetes.yaml')
k8s_resource('php-package-cache', port_forwards=['8080'], resource_deps=['setup-s3-bucket'])

target='prod'
live_update=[]
if os.environ.get('PROD', '') ==  '':
  target='build-env'
  live_update=[
    sync('go.mod', '/app/go.mod'),
    sync('go.sum', '/app/go.sum'),
    sync('pkg',    '/app/pkg'),
    sync('main.go', '/app/main.go'),
    run('go install .'),
  ]

docker_build(
  ref='ghcr.io/turbine-kreuzberg/php-package-cache:latest',
  context='.',
  dockerfile='deployment/Dockerfile',
  live_update=live_update,
  target=target,
  only=[ 'go.mod'
       , 'go.sum'
       , 'pkg'
       , 'main.go'
       , 'deployment/entrypoint.sh'
  ],
  ignore=[ '.git'
         , '*/*_test.go'
         , '*.yaml'
  ],
)
