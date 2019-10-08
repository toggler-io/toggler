# AWS Elastic Beanstalk

To deploy toggler to aws-eb, you can choose many different option.
The one I describe here is currently my preferred approach.
Here we will upload the compiled binary to the servers.

Before anything, you need to create an aws-eb application + environment.
After that, you need to set the `PORT` environment variable to `5000`.
This number came from the aws-eb documentation.
They use this value as default port.

Next we need to create the source bundle that can be uploaded.
To achieve this, you can find a builder script in the `bin` directory under the name `create-aws-eb-source`.

```bash
source .envrc
create-aws-eb-source
```

After this command is executed, you will find a `dist` folder with a `toggler.zip` file.
You can use this zip to upload the source in the aws-eb ui.

Alternatively, you may create a continuous deployment pipeline
where the tests of the project executed, and after a successful run
the `create-aws-eb-source` executed, and the `toggler.zip` file uploaded as the latest version.

To achieve this, you may find a aws deploy-policy template.
The output of the following command is the json file structure that you can paste into your aws `IAM` policy editor page.

```bash
ACCOUNT_ID="your-aws-account-id" envsubst <docs/deploy/aws/eb/deploy-policy.envsubst.json
```

## Amazon Relational Database Service (RDS)
if you use AWS RDS, and your environment is updated with the RDS environment variables,
you have to manually ad 1 more environment configuration, the `RDS_ENGINE`.
This variable helps the application figure out what adapter it should use to connect.

```bash
# example
RDS_ENGINE="postgres"
```
