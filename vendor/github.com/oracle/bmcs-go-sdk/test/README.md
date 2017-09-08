# Regression Tests

These are a series of console applications that will execute actual requests
against the Bare Metal API. Valid API credentials must be supplied in order for the
tests to succeed. The required credentials consist of a user OCID, the OCID of the
root tenancy of the user as well as the fingerprint of the user's RSA public key.
The caller must also supply a path to the user's private RSA key in pem format.


To run these regression tests:
1. Change to the baremetal-sdk-go project root
2. Make sure there is a .env file with the required credentials (There
   is a sample.env file located at the test directory)
3. Run the following command
```
make regression_test
```

See [API Docs](https://docs.us-phoenix-1.oraclecloud.com/) for more information on
authorization credentials.   


### Creating authentication for your .env file

* Generate an SSH key
```
ssh-keygen -t rsa -b 2048 -v # You MUST provide a passphrase
ssh-keygen -f oracle.pub -e -m PKCS8 # Keep this output, you'll paste into the Oracle console later
```

* Log into the Oracle console, hover over your email address in the top right, and click on "User Settings"
* Click "Add Public Key" and paste your output from the commands above into the console
* The Fingerprint value is displayed on the console once you add a key
* Grab the Tenancy ID from the bottom of any page
* CompartmentID you can get from the list show on the Identity->Compartments page of the  console. You may have to create a compartment.

