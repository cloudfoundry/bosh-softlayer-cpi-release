bosh-softlayer-cpi [![Build Status](https://travis-ci.org/maximilien/bosh-softlayer-cpi.svg?branch=master)](https://travis-ci.org/maximilien/bosh-softlayer-cpi#)
==================

An external [BOSH](http://github.com/cloudfoundry/bosh) CPI (cloud provider interface) for the [SoftLayer](http://www.softlayer.com) cloud and its [API](http://sldn.softlayer.com/article/SoftLayer-API-Overview)

Based on [cppforlife](https://github.com/cppforlife)'s [BOSH Warden CPI](https://github.com/cppforlife/bosh-warden-cpi)'s example external Go-language CPI.

**NOTE**: This is _work in progress_. No guarantees are implied that this will be usable nor finish. Consider this code as prototype code.

## Short `godep` Guide
* If you ever import a new package `foo/bar` (after you `go get foo/bar`, so that foo/bar is in `$GOPATH`), you can type `godep save ./...` to add it to the `Godeps` directory. (You can also use `godep save` in the directory that `main.go` is located. In our case, that would be `main`. It's probably easier to just use `godep save ./...` consistently. I'm not sure if there are any downsides to doing this, but I don't think there are.) Remember for each dependency project, to make sure you're working on the branch you're relying on as a dependency. `godep save` will save the version of the project of whatever branch you're on! I assume, in most cases, this would be `master`.
* To restore dependencies from the `Godeps` directory, simply use `godep restore`. `restore` is the opposite of `save`.
* If you ever remove a dependency or a link becomes deprecated, the easiest way is probably to remove your entire `Godeps` directory and run `godep save ./...` again, after making sure all your dependencies are in your `$GOPATH`. Don't manually edit `Godeps.json`!
* To update an existing dependency, you can use `godep update foo/bar` or `godep update foo/...` (where `...` is a wildcard)
* The godep project readme is a pretty good resource: https://github.com/tools/godep
