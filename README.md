prometheus-sd-dcosl4lb
======================

# What is for?

Find prometheus exporters using DC/OS's L4LB domain name.

# How to

### Build

```bash
# make docker
```

### Use

Deploy docker image on DC/OS. The network is should be "host" mode. This app scans `http://127.0.0.1:62080/v1/vips` endpoint, finds VIP domain name that starts with prefix: `prometheus`, collects their container ip and port, then writes JSON file for prometheus's `file_sd`.

```bash
$ prometheus-sd-dcosl4lb-linux -h
usage: sd adapter usage [<flags>]

Tool to generate file_sd target files for DC/OS L4LB SD mechanisms.

Flags:
  -h, --help                Show context-sensitive help (also try --help-long and --help-man).
      --l4lb.prefix="prometheus"
                            Prefix of DC/OS L4LB domain name for prometheus exporters. e.g.
                            <l4lb.prefix>.test.marathon.l4lb.thisdcos.directory.
      --l4lb.framework="*"  Framework name part of DC/OS L4LB domain name. e.g.
                            prometheus.test.<l4lb.framework>.l4lb.thisdcos.directory. To capture any frameworks, use "*" for name.
      --output.file="custom_sd.json"
                            Output file for file_sd compatible file.
      --listen.address="localhost:62080"
                            The address the DC/OS L4LB HTTP API is listening on for requests.
```

* The result JSON file example: 

```json
[
    {
        "targets": [
            "10.4.22.11:80",
            "10.4.94.3:80"
        ],
        "labels": {
            "__address__": "prometheus-testpromsd-test.marathon.l4lb.thisdcos.directory:80",
            "framework_name": "marathon",
            "service_id": "prometheus-testpromsd-test"
        }
    }
]
```

# Acknowledgement

* inspired by this blog post: https://prometheus.io/blog/2018/07/05/implementing-custom-sd/
