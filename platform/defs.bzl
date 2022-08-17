def _gba_rom_impl(ctx):
    program = ctx.file.program
    out = ctx.actions.declare_file(ctx.label.name + ".gba")
    title = ctx.attr.title
    if not title:
        title = ctx.label.name.upper()
    arguments = [
        "-title=" + title,
        program.path,
        out.path,
    ]
    ctx.actions.run(
        outputs = [out],
        inputs = [program],
        progress_message = "Creating ROM %s" % out.short_path,
        executable = ctx.executable._makerom,
        arguments = arguments,
    )
    return [DefaultInfo(files = depset([out]))]

gba_rom = rule(
    implementation = _gba_rom_impl,
    attrs = {
        "program": attr.label(
            mandatory = True,
            allow_single_file = True,
        ),
        "title": attr.string(),
        "_makerom": attr.label(
            default = Label("//tools/makerom"),
            allow_single_file = True,
            executable = True,
            cfg = "exec",
        ),
    },
)
