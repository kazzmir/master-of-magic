name := "Master of Magic"

version := "1.0"

scalaVersion := "2.9.1"

// sourceDirectories in Compile += file("src/MasterOfMagic/src")

// scalaSource in Compile <<= baseDirectory(_/"src/MasterOfMagic/src")

// scalaSource in Compile <<= baseDirectory(_/"src/MasterOfMagic/src")

unmanagedSourceDirectories in Compile <+= baseDirectory(_/"src/MasterOfMagic/src")

unmanagedSourceDirectories in Compile <+= baseDirectory(_/"src/Util/src")
