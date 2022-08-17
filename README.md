# Bazel GBA Example

This is an example project for building a Game Boy Advance game with Bazel.

## Caution

This downloads dependencies directly from the devKitPro package repository. Those packages will eventually be replaced with newer versions, and this repository will fail to build. At that point, you'll have to update this repo to a newer version of the packages.

Bazel will cache the downloads locally and share them across different projects. However, please be kind. Do not build this project inside a CI environment without using your own cache.

## Building

Requires Bazel. Tested on Linux and macOS. You do not have to install devKitARM, Bazel will download it automatically.

```shell
$ bazel build -c opt --platforms=//platform:gba //source:game
```

This will create the ROM image `bazel-bin/source/game.gba`. If you are developing, you should enable C compiler warnings and treat them as errors. You can do this by creating a `.bazelrc.user` file in the root workspace directory with the following line:

```
build --//bazel:warnings=error
```

The original ELF file will be placed at `bazel-bin/source/game.elf`. You can compile in debug mode by using `-c dbg` instead of `-c opt`. If you omit the `-c` option entirely, Bazel will simply build as fast as possible.

## Licensing

With some exceptions, code in the Bazel GBA example is licensed under the terms of the MIT license. See LICENSE.txt for details.

The file `platform/cc_toolchain_config.bzl` is licensed under the terms of the Apache License, version 2.0.

### LibGBA

LibGBA is open-source, licensed under the terms of the LGPL, with exceptions. 

See: https://github.com/devkitPro/libgba/blob/master/libgba_license.txt

### LibTonc

Note that libtonc is open-source, but the license is not documented in the libtonc Git repository.

See: https://web.archive.org/web/20200918104259/https://www.coranac.com/projects/tonc/

> cearn on 2018-06-24 at 9:56 said:

> Damian, tonclib is MIT licenced. Or would have been if I hadn't forgotten to add the license file >_>
