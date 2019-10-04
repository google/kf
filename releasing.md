**Before Release**
- [ ] Open a “Release vX.Y.Z” Issue and paste this checklist into the Issue body.

**Prepare Release**
- [ ] Fork a release branch off `develop` named `release-vX.Y.Z`
- [ ] Update the [Unreleased] anchor in CHANGELOG.md to the new version (in semver format). If the changelog seems incomplete, scan closed PRs for the milestone and add relevant “Release Notes” entries to it
- [ ] Update the `version` file to the new version (X.Y.Z)
- [ ] Push the release branch to your fork

**Github**
- [ ] Open a “Release vX.Y.Z” PR against the `master` branch
- [ ] Add the label “release”. This will trigger the Concourse release pipeline.
- [ ] Monitor the PR for feedback from Concourse
- [ ] Confirm that automated test, build, and publish steps have completed

**Complete Release**
- [ ] Merge the PR (“Merge Commit”) when the “OK to merge” comment is added by Concourse
- [ ] Create a new release (https://github.com/google/kf/releases).
  - [ ] “Tag version” is `vX.Y.Z`
  - [ ] “Target” is the release merge commit in master
  - [ ] Use the CHANGELOG.md contents as the release body
  - [ ] Attach the artifacts (manifests, licenses, and binaries) uploaded to GCS by Concourse. You can find the location of these artifacts in the "Publish" task of the "build-and-integrate" job in the pipeline.
- [ ] Publish the release

**Cleanup**
- [ ] Merge the release branch to the develop branch with the `--no-ff` flag
- [ ] Push the develop branch
- [ ] Delete the `release-vX.Y.Z` branch in the remote

