# Icon

## How to set the executable icon
I decided to use [rsrc](https://github.com/akavel/rsrc) to generate a syso file which is automatically embed
when building the executable to set the executable's icon.

To install rsrc you can use

```shell
go install github.com/akavel/rsrc@v0.10.2
```

Then use is it like this

```shell
rsrc -ico ./cmd/monitor.ico -o ./cmd/focusframe.syso
```

Or easier use

```shell
make syso
```
