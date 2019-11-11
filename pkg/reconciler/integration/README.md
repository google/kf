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

## Algorithms

The following are suggested algorithms to implement the validation tests.

### Defaulting Validation Algorithm

1. Create an object
2. Ensure no additional fields have been populated on the response in the
   `spec` or parts of `metadata`.

### Health Validation Algorithm

1. Create an object
2. Watch each update until a successful condition has been met. Success should
   be defined by Rego expression.
3. Run Rego expression(s) for each update we receive. If an expression fails,
   the whole test should fail. These expressions SHOULD test that our
   interpretation of the status as a state machine is correct.
4. If some timeout occurs before the resource is ready, fail the test.

### Update Validation Algorithm

This is the same as the health validation algorithm, but after the initial
success, the `spec` should be modified and the test will run the health
validation again.

### Upgrade Validation Algorithm

1. With an object, ask Kuberntes to upgrade it from version X to Y to produce Z
2. Ask Kubernetes to downgrade Z from version Y to X.
3. Assert that the specs are unchanged and that both upgrade and downgrade
   were successful.
