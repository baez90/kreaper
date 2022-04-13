# -*- mode: Python -*-

local_resource(
    'build',
    'CGO_ENABLED=0 go build -trimpath -ldflags "-w -s" -installsuffix cgo -o dist/kreaper main.go',
    deps=['main.go', './reaper'],
)

debug_dockerfile = """
FROM docker.io/alpine:3.15

COPY kreaper /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/kreaper"]
"""

custom_build(
    'kreaper',
    'docker build -f deployments/Dockerfile -t $EXPECTED_REF --build-arg BASE="docker.io/alpine:3.15" ./dist/',
    entrypoint='/usr/local/bin/kreaper',
    deps=['./dist/kreaper'],
    live_update=[
        sync('./dist/kreaper', '/usr/local/bin/kreaper'),
    ]
)

k8s_yaml(['testdata/target_pod.yaml', 'testdata/deployment.yaml'])
k8s_resource('kreaper', resource_deps=['build'])