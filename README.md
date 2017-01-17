bosh-softlayer-cpi [![Build Status](https://travis-ci.org/cloudfoundry/bosh-softlayer-cpi.svg?branch=master)](https://travis-ci.org/cloudfoundry/bosh-softlayer-cpi#)
==================

An external [BOSH](http://github.com/cloudfoundry/bosh) CPI (cloud provider interface) for the [SoftLayer](http://www.softlayer.com) cloud and its [API](http://sldn.softlayer.com/article/SoftLayer-API-Overview)

## Getting Started (*)
------------------

TBD

### Overview Presentations (*)
--------------------------

1. CF Summit 05/12/2015: [bosh-init](https://github.com/maximilien/presentations/blob/master/2015/cf-summit-bosh/releases/bosh-init-v1.0.0.pdf) - which also includes some details around BOSH external CPIs (motivation and goals)

### Cloning and Building
------------------------

Clone this repo and build it. Using the following commands on a Linux or Mac OS X system:

```
$ mkdir -p bosh-softlayer-cpi/src/github.com/cloudfoundry
$ export GOPATH=$(pwd)/bosh-softlayer-cpi:$GOPATH
$ cd bosh-softlayer-cpi/src/github.com/cloudfoundry
$ git clone https://github.com/cloudfoundry/bosh-softlayer-cpi.git
$ cd bosh-softlayer-cpi
$ ./bin/build
$ ./bin/test-unit
$ export SL_USERNAME=your-username@your-org.com
$ export SL_API_KEY=your-softlayer-api-key
$ ./bin/test-integration
```

NOTE: if you get any dependency errors, then use `go get path/to/dependency` to get it, e.g., `go get github.com/onsi/ginkgo/ginkgo` and `go get github.com/onsi/gomega`

The executable output should now be located in: `out/cpi`. You will need to package this into a BOSH release. The easiest way is to use the [bosh-softlayer-cpi-release](https://github.com/cloudfoundry-incubator/bosh-softlayer-cpi-release) project.

### Running Tests
-----------------

The [SoftLayer](http://www.softlayer.com) (SL) CPI and associated tests and binary distribution depend on you having a real SL account. Get one for free for one month [here](http://www.softlayer.com/info/free-cloud). From your SL account you can get an API key. Using your account name and API key you will need to set two environment variables: `SL_USERNAME` and `SL_API_KEY`. You can do so as follows:

```
$ export SL_USERNAME=your-username@your-org.com
$ export SL_API_KEY=your-softlayer-api-key
```

You should run the tests to make sure all is well, do this with: `$ ./bin/test-unit` and `$ ./bin/test-integration` in your cloned repository. Please note that the `$ ./bin/test-integration` will spin up real SoftLayer virtual guests (VMs) and associated resources and will also delete them. This integration test may take up to 30 minutes (usually shorter)

The output should of `$ ./bin/test-unit` be similar to:

```
$ ./bin/build
$ ./bin/test-unit

 Cleaning build artifacts...

 Formatting packages...

 Integration Testing packages:
Will skip:
  ./integration
  ./integration/has_vm
  ./integration/set_vm_metadata
[1435781395] Action Suite - 80/80 specs - 7 nodes •••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••••• SUCCESS! 39.715515ms
[1435781395] Dispatcher Suite - 19/19 specs - 7 nodes ••••••••••••••••••• SUCCESS! 12.016828ms
[1435781395] Transport Suite - 3/3 specs - 7 nodes ••• SUCCESS! 3.195291ms
[1435781395] Common Suite - 0/0 specs - 7 nodes  SUCCESS! 4.824792ms
[1435781395] Main Suite - 8/8 specs - 7 nodes •••••••• SUCCESS! 12.886684ms
[1435781395] Baremetal Suite - 10/10 specs - 7 nodes •••••••••• SUCCESS! 7.567213ms
[1435781395] Disk Suite - 6/6 specs - 7 nodes •••••• SUCCESS! 26.963978ms
[1435781395] Stemcell Suite - 4/4 specs - 7 nodes •••• SUCCESS! 8.112732ms
[1435781395] VM Suite - 39/39 specs - 7 nodes ••••••••••••••••••••••••••••••••••••••• SUCCESS! 16.783455ms
[1435781395] TestHelpers Suite - 6/6 specs - 7 nodes •••••• SUCCESS! 8.888439ms

Ginkgo ran 10 suites in 2.341344604s
Test Suite Passed

 Vetting packages for potential issues...

SWEET SUITE SUCCESS
```

## Developing
-------------

1. Check for existing stories on our [public Tracker](https://www.pivotaltracker.com/n/projects/1344876)
2. Select an unstarted story and work on code for it
3. If the story you want to work on is not there then open an issue and ask for a new story to be created
4. Run `go get golang.org/x/tools/cmd/vet`
5. Run `go get github.com/xxx ...` to install test dependencies (as you see errors)
6. Write a [Ginkgo](https://github.com/onsi/ginkgo) test.
7. Run `bin/test` and watch the test fail.
8. Make the test pass.
9. Submit a pull request.

## Contributing
---------------

* We gratefully acknowledge and thank the [current contributors](https://github.com/cloudfoundry/bosh-softlayer-cpi/graphs/contributors)
* We welcome any and all contributions as Pull Requests (PR)
* We also welcome issues and bug report and new feature request. We will address as time permits
* Follow the steps above in Developing to get your system setup correctly
* Please make sure your PR is passing Travis before submitting
* Feel free to email me or the current collaborators if you have additional questions about contributions
* Before submitting your first PR, please read and follow steps in [CONTRIBUTING.md](CONTRIBUTING.md)

### Managing dependencies
-------------------------

* All dependencies managed via [Godep](https://github.com/tools/godep). See [Godeps/_workspace](https://github.com/cloudfoundry/bosh-softlayer-cpi/tree/master/Godeps/_workspace) directory on master

#### Short `godep` Guide
* If you ever import a new package `foo/bar` (after you `go get foo/bar`, so that foo/bar is in `$GOPATH`), you can type `godep save ./...` to add it to the `Godeps` directory.
* To restore dependencies from the `Godeps` directory, simply use `godep restore`. `restore` is the opposite of `save`.
* If you ever remove a dependency or a link becomes deprecated, the easiest way is probably to remove your entire `Godeps` directory and run `godep save ./...` again, after making sure all your dependencies are in your `$GOPATH`. Don't manually edit `Godeps.json`!
* To update an existing dependency, you can use `godep update foo/bar` or `godep update foo/...` (where `...` is a wildcard)
* The godep project [readme](https://github.com/tools/godep/README.md) is a pretty good resource: [https://github.com/tools/godep](https://github.com/tools/godep)

* Since GO1.5, dependencies can be managed via [Govendor](https://github.com/kardianos/govendor). See [vendor](https://github.com/cloudfoundry/bosh-softlayer-cpi/tree/master/vendor) directory.

### Current conventions
-----------------------

* Basic Go conventions
* Strict TDD for any code added or changed
* Go fakes when needing to mock objects

## Credit
---------

The base code was inpired by [cppforlife](https://github.com/cppforlife)'s [BOSH Warden CPI](https://github.com/cppforlife/bosh-warden-cpi)'s example external Go-language CPI.

**NOTE**: This is still _work in progress_. No guarantees are implied that this will be usable nor finish. Consider this code as prototype code.

(*) these items are in the works, we will remove the * once they are available
