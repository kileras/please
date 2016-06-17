package(default_visibility = ['PUBLIC'])

maven_jar(
    name = 'junit',
    id = 'junit:junit:4.12',
    hash = 'sha1: a791201ac8a3d2a251045a52e264de01343ad2df',
)

maven_jar(
    name = 'hamcrest',
    id = 'org.hamcrest:hamcrest-all:1.3',
    hash = 'sha1: 6e72502584462b2d35a7a6a18fc5541c412f08cc',
)

maven_jars(
    name = 'guava',
    id = 'com.google.guava:guava:18.0',
    exclude = ['jsr305'],
    hashes = {
        'guava': 'sha1: 8908f57125132a4468ea9977705a83f61b94fa89',
    }
)

maven_jars(
    name = 'jacoco',
    id = 'org.jacoco:jacoco:0.7.6.201602180812',
    exclude = [
        'org.jacoco.ant',
        'org.jacoco.doc',
        'org.jacoco.examples',
        'asm-debug-all',
    ],
    hashes = {
        'org.jacoco.report': 'sha1: fe79a683193a3e897e734add5abae9af27eedf13',
        'org.jacoco.agent': 'sha1: 536da5bb74368a21bf322dff9413c23c7b595b61',
        'org.jacoco.core': 'sha1: ecd020a32fdfc96c6c845422978b362e11bdde5d',
    },
    deps = [':asm'],
    deps_only = True,
)

maven_jar(
    name = 'asm',
    id = 'org.ow2.asm:asm-debug-all:5.0.4',
    hash = 'sha1: 079196a08a86094763068ac621ba90c73c2efa34',
)

maven_jar(
    name = 'logback-classic',
    id = 'ch.qos.logback:logback-classic:1.1.3',
    hash = 'sha1: 245ace75b7a18931f103a79e74d53c66f1f90a9e',
    deps = [
        ':logback-core',
        ':slf4j-api',
    ]
)

maven_jar(
    name = 'logback-core',
    id = 'ch.qos.logback:logback-core:1.1.3',
    hash = 'sha1: 53732a6ab1f058ad76808646519801e30cc4a780',
)

maven_jar(
    name = 'slf4j-api',
    id = 'org.slf4j:slf4j-api:1.7.10',
    hash = 'sha1: 0ff5ea6b9f644c97ef1de83eb75e7cc304103549',
)

maven_jar(
    name = 'protobuf',
    id = 'com.google.protobuf:protobuf-java:3.0.0-beta-1',
    hash = 'b832e95bd94c68a01c32985d99ca5bb74bb0e0dc',
)

github_file(
    name = 'bazel-deps-src',
    repo = 'pgr0ss/bazel-deps',
    file = 'src/main/java/braintree/Maven.java',
    revision = 'c9d7afea5b38a90db0358a2dfc5a4729ceaf4bd9',
    hash = 'd0fcb1fee6c4db1751c52bee4f8b5cd1d0a27f64',
)

java_library(
    name = 'bazel-deps',
    srcs = [':bazel-deps-src'],
    deps = [
        ':guava',
        ':aether-api',
        ':aether-impl',
        ':aether-connector-basic',
        ':aether-transport-file',
        ':aether-transport-http',
        ':maven-aether-provider',
    ],
)

# These are dependencies of please_maven, so of course we can't fetch them with maven_jars :(
maven_jar(
    name = 'aether-api',
    id = 'org.eclipse.aether:aether-api:1.0.2.v20150114',
    hash = '',
)
maven_jar(
    name = 'aether-impl',
    id = 'org.eclipse.aether:aether-impl:1.0.2.v20150114',
    hash = '',
)
maven_jar(
    name = 'aether-connector-basic',
    id = 'org.eclipse.aether:aether-connector-basic:1.0.2.v20150114',
    hash = '',
)
maven_jar(
    name = 'aether-transport-file',
    id = 'org.eclipse.aether:aether-transport-file:1.0.2.v20150114',
    hash = '',
)
maven_jar(
    name = 'aether-transport-http',
    id = 'org.eclipse.aether:aether-transport-http:1.0.2.v20150114',
    hash = '',
)
maven_jar(
    name = 'maven-aether-provider',
    id = 'org.apache.maven:maven-aether-provider:3.1.0',
    hash = '',
)
maven_jar(
    name = 'args4j',
    id = 'args4j:args4j:2.32',
    hash = '',
)
