# prometheus-vmware-exporter
Collect VMWare metrics from a vSphere host

## Test
Currently, the only tests available are benchmarks.  To run them, set
`VSPHERE_HOST`, `VSPHERE_USERNAME`, and `VSPHERE_PASSWORD` in your shell
environment and run `go test -bench . -benchmem` from the `controller`
directory.

## Build
```sh 
docker build -t prometheus-vmware-exporter .
```

## Run
```sh
docker run -d \
  --restart=always \
  --name=prometheus-vmware-exporter \
  --env=VSPHERE_HOST vsphere.domain.local \
  --env=VSPHERE_USERNAME user \
  --env=VSPHERE_PASSWORD password \
  --env=LOG_LEVEL debug \
  prometheus-vmware-exporter 
```
