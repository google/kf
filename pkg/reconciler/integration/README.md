## Kf Black Box Tests

The goal of these black box tests is to be able to test specific K8s types
without any knowledge about what they do; basically compliance tests.

This package implements the following classes of tests (listed in order of importance):

1. Defaulting validation: we ensure the defaults for a resource don't change by
   creating it and comparing the values. This will catch unexpected changes in
  defaulting webhooks. (This bitten us in 0.1 and 0.2)
2. Health validation: we ensure the status of the object behaves consistently
   with our expectations. (Ongoing issue)
3. Update validation: ensure that the status of updated objects behaves in
   accordance with our expectations.
4. Upgrade validation: ensure that upgrading an object from one version to the
   next goes as anticipated. Ideally this will be well-tested by the creators.

NOTE: These black-box-tests are only intended to run positive examples, if we
start expecting certain error messages, codes, etc. we're likely going to end up
on the un-maintainable spectrum.
