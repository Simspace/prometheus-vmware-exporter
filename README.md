# prometheus-vmware-exporter
Collect VMWare metrics from a vSphere host

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
