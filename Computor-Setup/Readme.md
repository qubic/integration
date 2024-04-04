# Qubic Computor Setup

> WIP

## EFI Baremetal Network Setup
In some cases your node will not start hanging at "Init TCP...". In most cases this means that your Server has multiple network interfaces and qubic tries to use the wrong one.

See [Disconnect-Unneeded-Devices](Disconnect-Unneeded-Devices.md) how to disconnect network devices you don't need.

## Troubleshooting

### EFI Protocol Error "reads invalid number of bytes"
If you see this error, this could mean that your node couldn't read/write from/to your disk (usb/nvme).

In the file [file_io.h](https://github.com/qubic/core/blob/main/src/platform/file_io.h) the read/write chunk size can be adjusted.

```c++
// If you get an error reading and writing files, set the chunk sizes below to
// the cluster size set for formatting you disk. If you have no idea about the
// cluster size, try 32768.
#define READING_CHUNK_SIZE 1048576
#define WRITING_CHUNK_SIZE 1048576
```
