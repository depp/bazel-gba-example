_BASE_URLS = [
    "https://downloads.devkitpro.org/packages/",
]

def devkitarm_urls(url):
    return [base + url for base in _BASE_URLS]

def _archive(url, sha256):
    return struct(
        url = url,
        sha256 = sha256,
    )

_TOOLCHAINS = {
    "linux.amd64": _archive(
        "linux/x86_64/devkitARM-r58-2-x86_64.pkg.tar.xz",
        "247f81d86f223d6a1c8e4288eda823cca0d3c5ed3327b1ff42473b01e482d631",
    ),
    "darwin.amd64": _archive(
        "osx/x86_64/devkitARM-r58-2-x86_64.pkg.tar.xz"
        "65564898ea485c92db52cc1beea90fc8ccada0e9a29873913684bcfb15d4b26b",
    ),
}

_ARCH_NORMALIZE = {
    "x86_64": "amd64",
}

def _get_toolchain(os):
    os_name = os.name
    arch = os.arch
    arch_norm = _ARCH_NORMALIZE.get(arch, arch)
    if os_name == "linux":
        keys = ["linux.{}".format(arch_norm)]
    elif os_name == "mac os x":
        if arch_norm == "aarch64":
            keys = ["darwin.aarch64", "darwin.amd64"]
        else:
            keys = ["darwin.{}".format(arch_norm)]
    else:
        fail("no toolchain available for OS: {}".format(os_name))
    for k in keys:
        t = _TOOLCHAINS.get(k)
        if t:
            return t
    fail("no toolchain available for architecture: {}".format(arch))

def _devkitarm_repository(ctx):
    toolchain = _get_toolchain(ctx.os)
    ctx.download_and_extract(
        devkitarm_urls(toolchain.url),
        "",
        toolchain.sha256,
        "",
        "opt/devkitpro/devkitARM",
    )
    ctx.file("WORKSPACE", "workspace(name = \"{}\")\n".format(ctx.name))
    ctx.file("BUILD.bazel", ctx.read(ctx.attr._build_file))

devkitarm_repository = repository_rule(
    implementation = _devkitarm_repository,
    attrs = {
        "_build_file": attr.label(
            default = Label("@//platform:devkitarm.bazel"),
        ),
    },
)
