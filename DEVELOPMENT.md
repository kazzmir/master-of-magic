# Development guidelines

# Workflow

Create a new branch for bug fixes or new features. There are no constraints on the name of the branch so long as the name is reasonable, and understandable by other developers. Branches can be pushed to github early on in the life of the branch, even if the feature is not done. Create a pull request in github and possibly mark it as a draft if the branch still needs work on it. Once the feature is done and passes the tests, the pull request will be merged back to master.

Branches are always squash merged back into master to keep linear history.

Some branch guidelines:
  * merge master into your branch frequently so that the branch does not diverge from master too much
  * try not to let the branch be too long lived. A maximum of a week is probably enough. If there is so much work that a branch takes longer than a week to finish, then try to split the branch into smaller chunks that can be merged in isolation.
  * look at other in-flight branches to see what changes might conflict with yours
  * not every issue and detail has to be fixed in a new feature. FIXME comments can be added and fixed in a future pull request

# Tests

Try to add tests to new features. Try to set up the code so that the majority of the functionality does not directly depend on having an lbx.LbxCache in order to run, so that the tests can run in github actions where the lbx files are not available.

# Code Style

 * gofmt is not used. Indentation is 4 spaces. Otherwise just try to match the surrounding code style with regard to bracket placement and indentation.
 * import order is roughly
```
import (
  <standard libraries, like io, log, etc>
  <master of magic modules>
  <3rd party libraries, ebiten, etc>
)
```
Within those three sets there is no constraint on the order in which imports appear.
 * variable names should be spelled out, no abbreviations/1 letter names except for a small set: x, y, z, i, j, err
 * avoid using panic, instead return an error

# Libraries

Try to avoid importing a new library unless absolutely necessary. In a lot of cases it might be simpler to vendor most of the library code into lib/
