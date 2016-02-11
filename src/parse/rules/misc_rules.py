"""Miscellaneous rules that aren't language-specific."""


def genrule(name, cmd, srcs=None, out=None, outs=None, deps=None, visibility=None,
            building_description='Building...', hashes=None, timeout=0, binary=False,
            needs_transitive_deps=False, output_is_complete=True, test_only=False,
            requires=None, provides=None, pre_build=None, post_build=None, tools=None):
    """A general build rule which allows the user to specify a command.

    Args:
      name (str): Name of the rule
      cmd (str): Command to run. It's subject to various sequence replacements:
             $(location //path/to:target) expands to the location of the given build rule, which
                                          must have a single output only.
             $(locations //path/to:target) expands to the locations of the outputs of the given
                                           build rule, which can have any number of outputs.
             $(exe //path/to:target) expands to a command to run the output of the given target.
                                     The rule must be marked as binary.
             $(out_location //path_to:target) expands to the output of the given build rule, with
                                              the preceding plz-out/gen etc.
           Also a number of environment variables will be defined:
             ARCH: architecture of the system, eg. amd64
             OS: current operating system (linux, darwin, etc).
             PATH: usual PATH environment variable as defined in your .plzconfig
             TMP_DIR: the temporary directory you're compiling within.
             SRCS: the sources of your rule
             OUTS: the outputs of your rule
             PKG: the path to the package containing this rule
             NAME: the name of this build rule
             OUT: the output of this rule. Only present when there is only one output.
             SRC: the source of this rule. Only present when there is only one source.
             SRCS_<suffix>: Present when you've defined named sources on a rule. Each group
                            creates one of these these variables with paths to those sources.
      srcs (list): Sources of this rule. Can be a list of files or rules, or a dict of names to similar
            lists. In the latter case they can be accessed separately which is useful when you
            have separate kinds of things in a rule.
      outs (list): Outputs of this rule.
      out (str): A single output of this rule, as a string. Discouraged in favour of 'outs'.
      deps (list): Dependencies of this rule.
      tools (list): Tools used to build this rule; similar to srcs but are not copied to the temporary
                    build directory. Should be accessed via $(exe //path/to:tool) or similar.
      visibility (list): Visibility declaration of this rule
      building_description (str): Description to display to the user while the rule is building.
      hashes (list): List of hashes; if given the outputs must match one of these. They can be
              optionally preceded by their method. Currently the only supported method is sha1.
      timeout (int): Maximum time in seconds this rule can run for before being killed.
      binary (bool): True to mark a rule that produces a runnable output. Its output will be placed into
              plz-out/bin instead of plz-out/gen and can be run with 'plz run'. Binary rules
              can only have a single output.
      needs_transitive_deps (bool): If True, all transitive dependencies of the rule will be made
                             available to it when it builds (although see below...). By default
                             rules only get their immediate dependencies.
      output_is_complete (bool): If this is true then the rule blocks downwards searches of transitive
                          dependencies by other rules (ie. it will be available to them, but not
                          its dependencies as well).
      test_only (bool): If True it can only be used by test rules.
      requires (list): A list of arbitrary strings that define kinds of output that this rule might want.
                See 'provides' for more detail; it's mostly useful to match up rules with multiple
                kinds of output with ones that only need one of them, eg. a proto_library with
                a python_library that doesn't want the C++ or Java proto outputs.
                Entries in 'requires' are also implicitly labels on the rule.
      provides (dict): A map of arbitrary strings to dependencies of the rule that provide some specific
                type of thing. For example:
                  provides = {'py': ':python_rule', 'go': ':go_rule'},
                A Python rule would have requires = ['py'] and so if it depended on a rule like
                this it would pick up a dependency on :python_rule instead. See the proto rules
                for an example of where this is useful.
                Note that the keys of provides and entries in requires are arbitrary and
                have no effect until a matched pair meet one another.
      pre_build (function): A function to be executed immediately before the rule builds. It receives one
                 argument, the name of the building rule. This is mostly useful to interrogate
                 the metadata of dependent rules which isn't generally available at parse time;
                 see the get_labels function for a motivating example.
      post_build (function): A function to be executed immediately after the rule builds. It receives two
                  arguments, the rule name and its command line output.
                  This is significantly more useful than the pre_build function, it can be used
                  to dynamically create new rules based on the output of another.
    """
    if out and outs:
        raise TypeError('Can\'t specify both "out" and "outs".')
    build_rule(
        name=name,
        srcs=srcs,
        outs=[out] if out else outs,
        cmd=cmd,
        deps=deps,
        tools=tools,
        visibility = visibility,
        output_is_complete=output_is_complete,
        building_description=building_description,
        hashes=hashes,
        post_build=post_build,
        binary=binary,
        build_timeout=timeout,
        needs_transitive_deps=needs_transitive_deps,
        requires=requires,
        provides=provides,
        test_only=test_only,
    )


def gentest(name, test_cmd, labels=None, cmd=None, srcs=None, outs=None, deps=None, tools=None,
            data=None, visibility=None, timeout=0, needs_transitive_deps=False, flaky=0,
            no_test_output=False, output_is_complete=True, requires=None, container=False):
    """A rule which creates a test with an arbitrary command.

    The command must return zero on success and nonzero on failure. Test results are written
    to test.results (or not if no_test_output is True).
    Most arguments are similar to genrule() so we cover them in less detail here.

    Args:
      name (str): Name of the rule
      test_cmd (str): Command to run for the test.
      labels (list): Labels to apply to this test.
      cmd (str): Command to run to build the test.
      srcs (list): Source files for this rule.
      outs (list): Output files of this rule.
      deps (list): Dependencies of this rule.
      tools (list): Tools used to build this rule; similar to srcs but are not copied to the temporary
                    build directory. Should be accessed via $(exe //path/to:tool) or similar.
      data (list): Runtime data files for the test.
      visibility (list): Visibility declaration of this rule.
      timeout (int): Length of time in seconds to allow the test to run for before killing it.
      needs_transitive_deps (bool): True if building the rule requires all transitive dependencies to
                             be made available.
      flaky (bool | int): If true the test will be marked as flaky and automatically retried.
      no_test_output (bool): If true the test is not expected to write any output results, it's only
                      judged on its return value.
      output_is_complete (bool): If this is true then the rule blocks downwards searches of transitive
                          dependencies by other rules.
      requires (list): Kinds of output from other rules that this one requires.
      container (bool | dict): If true the test is run in a container (eg. Docker).
    """
    build_rule(
        name=name,
        srcs=srcs,
        outs=outs,
        deps=deps,
        data=data,
        tools=tools,
        test_cmd = test_cmd,
        cmd=cmd or 'true',  # By default, do nothing
        visibility=visibility,
        output_is_complete=output_is_complete,
        labels=labels,
        binary=True,
        test=True,
        test_timeout=timeout,
        needs_transitive_deps=needs_transitive_deps,
        requires=requires,
        container=container,
        no_test_output=no_test_output,
        flaky=flaky,
    )


def export_file(name, src, visibility=None, binary=False, test_only=False):
    """Essentially a single-file alias for filegroup.

    Args:
      name (str): Name of the rule
      src (str): Source file for the rule
      visibility (list): Visibility declaration
      binary (bool): True to mark the rule outputs as binary
      test_only (bool): If true the exported file can only be used by test targets.
    """
    filegroup(
        name = name,
        srcs = [src],
        visibility = visibility,
        binary = binary,
        test_only = test_only,
    )


def filegroup(name, srcs=None, deps=None, exported_deps=None, visibility=None, labels=None, binary=False,
              output_is_complete=True, requires=None, provides=None, link=True, test_only=False):
    """Defines a collection of files which other rules can depend on.

    Sources can be omitted entirely in which case it acts simply as a rule to collect other rules,
    which is often more handy than you might think.

    Args:
      name (str): Name of the rule
      srcs (list): Source files for the rule.
      deps (list): Dependencies of the rule.
      exported_deps (list): Dependencies that will become visible to any rules that depend on this rule.
      visibility (list): Visibility declaration
      labels (list): Labels to apply to this rule
      binary (bool): True to mark the rule outputs as binary
      output_is_complete (bool): If this is true then the rule blocks downwards searches of transitive
                                 dependencies by other rules.
      requires (list): Kinds of output from other rules that this one requires.
      provides (dict): Kinds of output that this provides for other rules (see genrule() for a more
                       in-depth discussion of this).
      test_only (bool): If true the exported file can only be used by test targets.
    """
    cmd = 'true'  # Nothing to do if we only have deps.
    if srcs:
        not_root = bool(get_base_path())
        locations = ' '.join('$(location_pairs %s)' % src for src in srcs
                             if not_root or src.startswith(':') or src.startswith('/'))
        if locations:
            cmd = 'echo %s | xargs -n 2 %s' % (locations, 'ln -s' if link else 'cp -r')
    build_rule(
        name=name,
        srcs=srcs,
        deps=deps,
        exported_deps=exported_deps,
        outs=srcs,
        cmd=cmd,
        visibility=visibility,
        building_description='Symlinking...' if link else 'Copying...',
        # This fixes some issues; I think it's reasonable that the outputs of filegroups
        # are treated just as files without any transitive deps.
        output_is_complete=output_is_complete,
        # This just symlinks its inputs so it's faster not to copy to the cache and back,
        # especially if the files it's collecting are large.
        skip_cache=link,
        requires=requires,
        provides=provides,
        test_only=test_only,
        labels=labels,
        binary=binary,
    )


def remote_file(name, url, hashes, out=None, binary=False, visibility=None, test_only=False):
    """Defines a rule to fetch a file over HTTP(S).

    Args:
      name (str): Name of the rule
      url (str): URL to fetch
      hashes (list): List of hashes; the output must match at least one of these. This is required
                     because the remote file must not change, otherwise it'd introduce fundamental
                     indeterminacy into the build.
      out (str): Output name of the file. Chosen automatically if not given.
      binary (bool): True to mark the output as binary and runnable.
      visibility (list): Visibility declaration of the rule.
      test_only (bool): If true the rule is only visible to test targets.
    """
    cmd = 'curl %s -o %s' % (url, out) if out else 'curl %s -O' % url
    # TODO(pebers): maybe plz should automatically do this on binary outputs?
    if binary:
        cmd += ' && chmod +x $OUT'
    build_rule(
        name=name,
        cmd=cmd,
        outs=[out or url[url.rfind('/') + 1:]],
        binary=binary,
        visibility=visibility,
        hashes=hashes,
        building_description='Fetching...',
    )


def fpm_package(name, files, version, package_type, links=None, package_name=None, options='',
                srcs=None, deps=None, visibility=None, labels=None):
    """Defines a rule to build a package using fpm.

    Args:
      name (str): Rule name
      files (dict): Dict of locations -> files to include, for example:
             {
                 '/usr/bin/plz': '//src:please',
                 '/usr/share/plz/junit_runner': '//src/build/java:junit_runner',
                 '/usr/share/plz/some_file': 'some_file',  # file in this package
             }
      links (dict): Dict of locations -> file to link to, for example:
             {
                 '/usr/bin/plz': '/opt/please',
             }
      version (str): Version of the package.
      package_type (str): Type of package to build (deb, rpm, etc)
      package_name (str): Name of package. Defaults to rule name.
      options (str): Extra options to pass to fpm.
      srcs (list): Extra sources (it's not necessary to mention entries in 'files' here)
      deps (list): Dependencies
      visibility (list): Visibility specification.
      labels (list): Labels associated with this rule.
    """
    package_name = package_name or name
    cmd = ' && '.join('mkdir -p $(dirname %s) && cp -r ../$(location %s) %s' %
                      (k.lstrip('/'), v, k.lstrip('/')) for k, v in sorted(files.items()))
    if links:
        cmd += ' && ' + ' && '.join('mkdir -p $(dirname %s) && ln -s %s %s' %
                                    (k.lstrip('/'), v, k.lstrip('/')) for k, v in sorted(links.items()))
    cmd = 'mkdir _tmp && cd _tmp && %s && fpm -s dir -t %s -n "%s" -v "%s" %s -p $OUT .' % (
        cmd, package_type, package_name, version, options)
    build_rule(
        name=name,
        srcs=sorted(files.values()) + (srcs or []),
        outs=['%s_%s_%s.deb' % (package_name, version, CONFIG.ARCH)],
        cmd=cmd,
        deps=deps,
        visibility=visibility,
        building_description='Packaging...',
        requires=['fpm'],
    )


def fpm_deb(name, files, version, links=None, package_name=None, options='',
            srcs=None, deps=None, visibility=None, labels=None):
    """Convenience wrapper around fpm_package that always builds a .deb package.

    Args:
      name (str): Rule name
      files (dict): Dict of locations -> files to include, for example:
             {
                 '/usr/bin/plz': '//src:please',
                 '/usr/share/plz/junit_runner': '//src/build/java:junit_runner',
                 '/usr/share/plz/some_file': 'some_file',  # file in this package
             }
      links (dict): Dict of locations -> file to link to, for example:
             {
                 '/usr/bin/plz': '/opt/please',
             }
      version (str): Version of the package.
      package_name (str): Name of package. Defaults to rule name.
      options (str): Extra options to pass to fpm.
      srcs (list): Extra sources (it's not necessary to mention entries in 'files' here)
      deps (list): Dependencies
      visibility (list): Visibility specification.
      labels (list): Labels associated with this rule.
    """
    fpm_package(
        name=name,
        files=files,
        version=version,
        package_type='deb',
        links=links,
        package_name=package_name,
        options=options,
        srcs=srcs,
        deps=deps,
        visibility=visibility,
        labels=labels,
    )


def tarball(name, srcs, out=None, deps=None, subdir=None,
            compression='gzip', visibility=None, labels=None):
    """Defines a rule to create a tarball containing outputs of other rules.

    Args:
      name: Rule name
      srcs: Source files to include in the tarball
      out: Name of output tarball (defaults to `name`.tar.gz, but see below re compression)
      subdir: Subdirectory to create in (defaults to 'name')
      compression: Kind of compression to use. Either one of {gzip, bzip2, xz, lzma}
                   to filter through known tar methods, an explicit flag, or None for
                   no compression.
      deps: Dependencies
      visibility: Visibility specification.
      labels: Labels associated with this rule.
    """
    subdir = subdir or name
    locations = ' '.join('$(location_pairs %s)' % src for src in srcs)
    if compression is not None and compression.startswith('-'):
        if not out:
            raise ValueError('Must pass "out" argument to tarball() if you pass an '
                             'explicit flag for "compression"')
    else:
        compression, extension = _COMPRESSION.get(compression, ('-a', ''))
    build_rule(
        name=name,
        cmd=' && '.join([
            'mkdir -p _tmp/' + subdir,
            'cd _tmp/' + subdir,
            'echo %s | xargs -n 2 cp -r' % locations,
            'cd ${TMP_DIR}/_tmp',
            'tar %s -cf $OUT *' % compression,
        ]),
        srcs=srcs,
        outs=[out or name + '.tar' + extension],
        deps=deps,
        visibility=visibility,
        labels=(labels or []) + ['tar'],
    )


_COMPRESSION = {
    'gzip': ('-z', '.gz'),
    'bzip2': ('-j', '.bz2'),
    'xz': ('-J', '.xz'),
    'lzma': ('--lzma', '.lzma'),
    'compress': ('-Z', '.Z'),
}


def _tool_path(tool, tools=None):
    """Returns the invocation of a tool and the list of tools for a rule to depend on.

    Used for tools like pex_tool and jarcat_tool which might be repo rules or just filesystem paths.
    """
    if tool.startswith('//'):
        return '$(exe %s)' % tool, [tool] + (tools or [])
    return tool, tools