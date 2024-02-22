# Mikrotik Wifi Controller

This little program creates wifi networks (virtual access points) using the RouterOS API built into Mikrotik routers. It's a work in progress, for fun.

## Set up connection params

```bash
export MIKROTIK_ADDRESS=192.168.88.1
export MIKROTIK_USERNAME=admin
export MIKROTIK_PASSWORD=hunter2_obviously
export MIKROTIK_PORT=8728
```

Or use like so:

```bash
./mikrotik-wifi command -address=129.168.88.1 -post=8728 -username=admin -password=hunter2
```

## List all networks

```bash
./mikrotik-wifi list
```

## Create a new network

```bash
./mikrotik-wifi create my-new-wifi hunter2
```

## Change the SSID name

```bash
./mikrotik-wifi update ssid my-new-wifi my-even-better-wifi
```

## Change the password

```bash
./mikrotik-wifi update password my-new-wifi hunter3
```

## Remove a network

```bash
./mikrotik-wifi remove my-new-wifi
```