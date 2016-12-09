# Testing the CPI

To test the CPI, you can issue direct commands to the CPI mimicking how the BOSH director invokes the CPI. 

Assuming that you are in the CPI project and have built it successfully with `bin/build` and passed tests with `bin/test` then the general command is:

```
out/cpi -configPath dev/config.json < dev/<cpi_method>.json
```

The [`dev/<cpi_method>.json`](https://bosh-softlayer-cpi/tree/master/dev) files are ready for you to modify and reuse.

Please note that the [`dev/config.json`](https://bosh-softlayer-cpi/tree/master/dev/config.json) needs to be modified once to include your SoftLayer `username` and `apiKey` instead of the fake ones listed.
