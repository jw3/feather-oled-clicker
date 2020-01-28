clicker
===

clicker firmware and daemon

```
command: model.sh
concurrency: 1
items:
  - title: test 1
    modules:
      - id: xrd0
        model: '[{"x":0,"y":0,"z":0.0505},{"x":1,"y":0,"z":0.2}]'
      - id: xrd1
        model: '[{"x":0,"y":0,"z":0.2},{"x":1,"y":0,"z":0.0505}]'
  - title: test 2
    modules:
      - id: xrd0
        model: '[{"x":0,"y":0,"z":0.0505},{"x":1,"y":0,"z":0.0505}]'
```

### reference
- https://www.pcsuggest.com/run-shell-scripts-from-udev-rules/
- https://coreos.com/os/docs/latest/using-systemd-and-udev-rules.html
- https://www.linode.com/docs/quick-answers/linux/start-service-at-boot/
- http://henrysbench.capnfatz.com/henrys-bench/arduino-sensors-and-input/keyes-ky-040-arduino-rotary-encoder-user-manual/
