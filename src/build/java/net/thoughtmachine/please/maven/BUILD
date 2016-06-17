java_library(
    name = 'maven',
    srcs = glob(['*.java'], excludes=['*Test.java']),
    deps = [
        '//third_party/java:guava',
        '//third_party/java:aether-api',
        '//third_party/java:aether-spi',
        '//third_party/java:aether-impl',
        '//third_party/java:aether-util',
        '//third_party/java:aether-connector-basic',
        '//third_party/java:aether-transport-http',
        '//third_party/java:aether-transport-file',
        '//third_party/java:maven-aether-provider',
        '//third_party/java:args4j',
    ],
)
