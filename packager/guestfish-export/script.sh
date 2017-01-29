mkdir /v2c/disk
guestfish -a /input/input -m /dev/sda1:/ copy-out / /v2c/disk
