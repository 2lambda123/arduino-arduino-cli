Arduino CLI provides support for reproducible builds through the use of build profiles.

## Sketch project file `sketch.yaml` and build profiles.

A profile is a complete description of all the resources needed to build a sketch. Profiles are defined in a project
file called `sketch.yaml`. This file is in YAML format and may contain multiple profiles.

Each profile will define:

- The board FQBN
- The target core platform name and version (with the 3rd party platform index URL if needed)
- A possible core platform name and version, that is a dependency of the target core platform (with the 3rd party
  platform index URL if needed)
- The libraries used in the sketch (including their version)

The format of the file is the following:

```
profiles:
  <PROFILE_NAME>:
    notes: <USER_NOTES>
    fqbn: <FQBN>
    platforms:
      - platform: <PLATFORM> (<PLATFORM_VERSION>)
        platform_index_url: <3RD_PARTY_PLATFORM_URL>
      - platform: <PLATFORM_DEPENDENCY> (<PLATFORM_DEPENDENCY_VERSION>)
        platform_index_url: <3RD_PARTY_PLATFORM_DEPENDENCY_URL>
    libraries:
      - <LIB_NAME> (<LIB_VERSION>)
      - <LIB_NAME> (<LIB_VERSION>)
      - <LIB_NAME> (<LIB_VERSION>)

...more profiles here...
```

There is a 'profiles:' section containing all the profiles. Each field is self-explanatory, in particular:

- `<PROFILE_NAME>` is the profile identifier, it’s a user-defined field, and the allowed characters are alphanumerics,
  underscore `_`, dot `.`, and dash `-`
- `<PLATFORM>` is the target core platform identifier, for example, `arduino:avr` or `adafruit:samd`
- `<PLATFORM_VERSION>` is the target core platform version required
- `<3RD_PARTY_PLATFORM_URL>` is the index URL to download the target core platform (also known as “Additional Boards
  Manager URLs” in the Arduino IDE). This field can be omitted for the official `arduino:*` platforms.
- `<PLATFORM_DEPENDENCY>`, `<PLATFORM_DEPENDENCY_VERSION>`, and `<3RD_PARTY_PLATFORM_DEPENDENCY_URL>` contains the same
  information as `<PLATFORM>`, `<PLATFORM_VERSION>`, and `<3RD_PARTY_PLATFORM_URL>` but for the core platform dependency
  of the main core platform. These fields are optional.
- `<LIB_VERSION>` is the version required for the library, for example, `1.0.0`
- `<USER_NOTES>` is a free text string available to the developer to add comments
- `<DEFAULT_PROFILE_NAME>` is the profile used by default (more on that later)

A complete example of a `sketch.yaml` may be the following:

```
profiles:
  nanorp:
    fqbn: arduino:mbed_nano:nanorp2040connect
    platforms:
      - platform: arduino:mbed_nano (2.1.0)
    libraries:
      - ArduinoIoTCloud (1.0.2)
      - Arduino_ConnectionHandler (0.6.4)
      - TinyDHT sensor library (1.1.0)

  another_profile_name:
    notes: testing the limit of the AVR platform, may be unstable
    fqbn: arduino:avr:uno
    platforms:
      - platform: arduino:avr (1.8.4)
    libraries:
      - VitconMQTT (1.0.1)
      - Arduino_ConnectionHandler (0.6.4)
      - TinyDHT sensor library (1.1.0)

  tiny:
    notes: testing the very limit of the AVR platform, it will be very unstable
    fqbn: attiny:avr:ATtinyX5:cpu=attiny85,clock=internal16
    platforms:
      - platform: attiny:avr@1.0.2
        platform_index_url: https://raw.githubusercontent.com/damellis/attiny/ide-1.6.x-boards-manager/package_damellis_attiny_index.json
      - platform: arduino:avr@1.8.3
    libraries:
      - ArduinoIoTCloud (1.0.2)
      - Arduino_ConnectionHandler (0.6.4)
      - TinyDHT sensor library (1.1.0)

  feather:
    fqbn: adafruit:samd:adafruit_feather_m0
    platforms:
      - platform: adafruit:samd (1.6.0)
        platform_index_url: https://adafruit.github.io/arduino-board-index/package_adafruit_index.json
    libraries:
      - ArduinoIoTCloud (1.0.2)
      - Arduino_ConnectionHandler (0.6.4)
      - TinyDHT sensor library (1.1.0)
```

### Building a sketch

When a `sketch.yaml` file exists in the sketch, it can be leveraged to compile the sketch with the `--profile/-m` flag
in the `compile` command:

```
arduino-cli compile --profile nanorp
```

In this case, the sketch will be compiled using the core platform and libraries specified in the nanorp profile. If a
core platform or a library is missing it will be automatically downloaded and installed on the fly in a isolated
directory inside the data folder. The dedicated storage is not accessible to the user and is meant as a "cache" of the
resources used to build the sketch.

When using the profile-based build, the globally installed platforms and libraries are excluded from the compile and can
not be used in any way. In other words, the build is isolated from the system and will rely only on the resources
specified in the profile: this will ensure that the build is portable and reproducible independently from the platforms
and libraries installed in the system.