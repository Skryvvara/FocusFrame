# FocusFrame



## Usage

> Please note that the usage of this application can change multiple times during the early stage of development.

Currently, the values for the window dimension, position and offset are hardcoded to fit as a 2560x1440 window
in the middle of a 49" ultrawide monitor, to change these values you currently have to modify the variables in the
`window` package.

While the application is running press `F4` to add or remove an application to/from the list of managed applications.
The application should then be moved to the desired values. Managed applications are added to or removed from `config.toml`.

This is an open source alternative to [Windowed Borderless Gaming](https://westechsolutions.net/sites/WindowedBorderlessGaming/home)
consider using that software for a more feature-complete and stable state.

## Build

To build the application from source use the steps listed below.

All required sources should be included in this repo including all the dependencies. So no internet connection should be required.

Building for Windows on non Windows machines could fail because of missing DLLs. I haven't tested this yet.


### Windows

```sh
make
```

The executable is then found in the `bin` directory (e.g) `bin/focusframe.exe`.

### Linux

*currently not supported*

### BSD

*currently not supported*

### Mac

*currently not supported*

## ToDo

- [x] move window and remove borders (core functionality)
- [x] create initial systray mockup
- [x] create config file
- [x] load (read) config file
- [ ] create UI to edit config + write config file
- [x] move managed apps automatically
- [ ] add functionality for linux (low priority but planned)
- [ ] add functionality for bsd (low priority but planned)
- [ ] add functionality for mac (low priority but planned)

## Compatibility

Currently, this tool is only compatible with Windows, although most parts and libs 
should work across Windows, linux, bsd and mac, the logic to move the desired window and
hide the window border is highly dependent on Windows functionality.

## Contribution

I am open for feature requests or code contributions. Feel free to create issues and/or merge requests.
Note that although I'm open for feature requests, this is a one-man project I'm toying with in my spare time.

## Special thanks



## License

- MIT

This program is licensed under the MIT license, meaning do whatever you want with it but don't call me when it explodes.
Although I'm happy to provide fixes for bugs found in my code if there's time.
