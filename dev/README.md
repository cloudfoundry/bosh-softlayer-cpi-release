# Testing the CPI

There are two general strategies for testing a CPI.

## Normal testing

Generally the CPI can be tested by using the CPI release and deploying that using bosh lite using vagrant. See the [bosh-softlayer-cpi-release](http://github.com/maximilien/bosh-softlayer-cpi-release/tree/master/README.md) for details.

Once the release is successfully deployed, the CPI is then located and associated with a BOSH director which can use it. You can then test the CPI by executing BOSH commands. If you make changes to the CPI code then you can test these changes by pushing your changes to github and doing the following:

1. update the CPI code under `src/bosh-softlayer-cpi` using:
  ```
  cd src/bosh-softlayer-cpi
  git pull
  ```

2. update your vagrant VM using:
  ```
  vagrant provision
  ```

1. executing some bosh commands, for example, to test:
  1. `create_stemcell` method, using: 
    ```
    bosh upload stemcell path/to/your/stemcell
    ```
  
  2. `create_vm` method, using: 
    ```
    bosh deploy
    ``` 
    assuming you are using the [dummy release](https://github.com/pivotal-cf-experimental/dummy-boshrelease) and have     created a release for it. 
    
    You can reuse the following [example dummy manifest YML](https://github.com/maximilien/bosh-softlayer-cpi/tree/master/dev/dummy-manifest.yml) file for a simple deployment with a single VM.

## Manual testing

In this approach, no release or director is used. Instead you will issue direct command to the CPI mimicking how the BOSH director invokes the CPI. 

Assuming that you are in the CPI project and have built it successfully with `bin/build` and passed tests with `bin/test` then the general command is:

```
out/cpi -configPath dev/config.json < dev/<cpi_method>.json
```

The [dev/<cpi_method>.json](https://github.com/maximilien/bosh-softlayer-cpi/tree/master/dev) files are ready for you to modify and reuse.

Please note that the [dev/config.json](https://github.com/maximilien/bosh-softlayer-cpi/tree/master/dev/config.json) needs to be modified once to include your SoftLayer `username` and `apiKey` instead of the fake ones listed.
