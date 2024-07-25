# am2manager (SincoANN profiles)

Tools to manage **am2** and **am2data** files.

## am2converter

Convert am2 to am2data (and am2data to am2data).

### Installation

```sh
go install github.com/walterwanderley/am2manager/cmd/am2converter@latest
```

### Usage

```sh
am2converter -level=120 -gain-min=0 -gain-max=100 file.am2 > file.am2Data
```

