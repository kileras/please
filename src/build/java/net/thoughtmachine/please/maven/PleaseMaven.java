package net.thoughtmachine.please.maven;

import java.util.ArrayList;
import java.util.Collection;
import java.util.Comparator;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;

import com.google.common.collect.ImmutableSet;
import com.google.common.collect.Sets;

import org.eclipse.aether.artifact.Artifact;
import org.eclipse.aether.artifact.DefaultArtifact;
import org.eclipse.aether.collection.DependencyCollectionException;
import org.kohsuke.args4j.Argument;
import org.kohsuke.args4j.CmdLineException;
import org.kohsuke.args4j.CmdLineParser;
import org.kohsuke.args4j.Option;

public class PleaseMaven {

  @Option(name = "-r", usage = "Repository to read from")
  private String repository = "https://repo1.maven.org/maven2";

  @Option(name = "-e", usage = "Artifacts to exclude")
  private List<String> excludeArtifact = new ArrayList<>();

  @Option(name = "-o", usage = "Optional dependencies to fetch")
  private List<String> optionalDeps = new ArrayList<>();

  @Argument(usage = "<artifact id>")
  private List<String> artifactNames = new ArrayList<>();

  public void run(String[] args) throws Exception {
    CmdLineParser parser = new CmdLineParser(this);
    parser.parseArgument(args);

    if (artifactNames.isEmpty()) {
      System.out.print("Usage: java -jar please_maven.jar");
      parser.printSingleLineUsage(System.out);
      System.out.println();
      parser.printUsage(System.out);
      System.out.println(
        "\nExample: java -jar please_Maven.jar com.fasterxml.jackson.core:jackson-databind:2.5.0");
      System.exit(1);
    }

    Maven maven = new Maven(repository, optionalDeps);
    for (String artifactName : artifactNames) {
      Set<Artifact> artifacts = maven.transitiveDependencies(new DefaultArtifact(artifactName));
      for (Artifact artifact : artifacts) {
        System.out.printf("%s:%s:%s\n", artifact.getGroupId(), artifact.getArtifactId(), artifact.getVersion());
      }
    }
  }

  public static void main(String[] args) throws Exception {
    new PleaseMaven().run(args);
  }
}
