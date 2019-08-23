using https://github.com/milesp20/intel_nuc_led kernel module

```
sudo apt install build-essential linux-headers-$(uname -r) debhelper dkms
sudo make dkms-install
```

to load module run `sudo modprobe nuc_led`

to load on boot add `nuc_led` line to `/etc/modules`

NOTE: root privileges needed to write
