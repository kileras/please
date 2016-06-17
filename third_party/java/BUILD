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

# These are dependencies of please_maven, so of course we can't fetch them with maven_jars :(
maven_jar(
    name = 'aether-core',
    id = 'org.eclipse.aether:aether-core:1.0.2.v20150114',
    hash = '',
)
maven_jar(
    name = 'aether-api',
    id = 'org.eclipse.aether:aether-api:1.0.2.v20150114',
    hash = '968d9fa7fe069c4cff001f7a9b17f681e3fd655b',
)
maven_jar(
    name = 'aether-spi',
    id = 'org.eclipse.aether:aether-api:1.0.2.v20150114',
    hash = '',
)
maven_jar(
    name = 'aether-impl',
    id = 'org.eclipse.aether:aether-impl:1.0.2.v20150114',
    hash = 'f3a63e29cc1527ce725dc69eb2d377caa1e92f19',
)
maven_jar(
    name = 'aether-connector-basic',
    id = 'org.eclipse.aether:aether-connector-basic:1.0.2.v20150114',
    hash = 'db50828b7bf5649190d8c28c16e9dd1e470d52fd',
)
maven_jar(
    name = 'aether-transport-file',
    id = 'org.eclipse.aether:aether-transport-file:1.0.2.v20150114',
    hash = 'c5dcb38aa79419c58e001efc96ef4733e466b56a',
)
maven_jar(
    name = 'aether-transport-http',
    id = 'org.eclipse.aether:aether-transport-http:1.0.2.v20150114',
    hash = 'fee2fbe46c2d67c220613d388bc45c9d783982d0',
)
maven_jar(
    name = 'maven-aether-provider',
    id = 'org.apache.maven:maven-aether-provider:3.1.0',
    hash = '3287089d27eaad587ba199fc2778c84bc53aae91',
)
maven_jar(
    name = 'args4j',
    id = 'args4j:args4j:2.32',
    hash = 'a08fb9a7217df438872f27936058f594eab29396',
)
