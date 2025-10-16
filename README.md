# Export test results to Test Reports

[![Step changelog](https://shields.io/github/v/release/bitrise-steplib/step-custom-test-results-export?include_prereleases&label=changelog&color=blueviolet)](https://github.com/bitrise-steplib/step-custom-test-results-export/releases)

This Step is used to search for test results and export them for the **Deploy to Bitrise.io** Step which can export the results to the Test Reports.


<details>
<summary>Description</summary>

Not all testing Steps automatically export their results to the Test Reports. This Step is designed to export the results of those Steps that do not do this automatically. The Step creates the required `test-info.json` file and deploys the test results in the correct directory for export.

### Configuring the Step

To use this Step to export the results of a testing Step - for example, **Flutter Test** -, you need to make sure that:

- The Step knows which folder to look for your test results.
- The Step knows which file to look for.
- The Step exports to the correct directory.

1. In the **The name of the test** input, set the name you want to see in the Test Reports. Each exported results file has a separate tab with the name defined here.

2. In the **Test result base path** input, set the path where your test results can be found.

   This input determines where Bitrise will look for your test results. We recommend setting a folder here, though you can also set a specific file path.

3. In the **Test result search pattern** input, set a glob pattern that matches a single test result output file.

   This search pattern is used to search every file and folder of the provided base path which was set in the **Test result base path** input. If there is more than one match, you will receive a warning and the Step will export the first match.

   If you set a specific file path in the previous input, just set `*` here.

4. In the **Step's test result directory** input, make sure the path points to a valid directory.

   By default, do NOT modify this input’s value. It should be set to the $BITRISE_TEST_RESULT_DIR Env Var.

5. Make sure you have a Deploy to Bitrise.io Step in your Workflow. This step is responsible for deploying the exported files from the test result directory to the Test Reports.

### Troubleshooting

- You always need a **Deploy to Bitrise.io** Step in your Workflow to export test results to the Test Reports. If you can't see your results in Test Reports, check that you have the Step in your Workflow.

- If more than one file matches the pattern set in the **Test result search pattern** input, you will receive a warning. It should only match a single file in the base path.

- If either the **Test result base path** input or the **Test result search pattern** input isn't set correctly or if either is left empty, the Step will fail.

### Useful links

[Test reports](https://devcenter.bitrise.io/testing/test-reports/)
[Exporting custom tests to test reports](https://devcenter.bitrise.io/testing/exporting-to-test-reports-from-custom-script-steps/)

### Related Steps

[Android Unit Test](https://www.bitrise.io/integrations/steps/android-unit-test)
[Xcode Test for iOS](https://www.bitrise.io/integrations/steps/xcode-test)

</details>

## 🧩 Get started

Add this step directly to your workflow in the [Bitrise Workflow Editor](https://docs.bitrise.io/en/bitrise-ci/workflows-and-pipelines/steps/adding-steps-to-a-workflow.html).

You can also run this step directly with [Bitrise CLI](https://github.com/bitrise-io/bitrise).

## ⚙️ Configuration

<details>
<summary>Inputs</summary>

| Key | Description | Flags | Default |
| --- | --- | --- | --- |
| `test_name` | Test name displayed in the Test Reports. | required |  |
| `base_path` | You can provide a path to a single file or a directory path and filter for files using the **Test result search pattern** input.  This is the base path where the Step will look for your test results. We recommend using a relative path to the results:  `app/build/test-results/testDemoDebugUnitTest/`  or  `./app/build/test-results/testDemoDebugUnitTest/`  | required | `$BITRISE_SOURCE_DIR` |
| `search_pattern` | This glob should match a single file within the provided base path. If it matches multiple files, the Step will produce a warning displaying all matches. The matched file should be in a supported format. If you set a specific file path in the **Test result base path** input, set this value to `*`.  **Examples**:  Matching all files within the base path: `*`  Matching all files within a given directory of the base path: `*/build/test-results/testDemoDebugUnitTest/*`  | required |  |
| `bitrise_test_result_dir` | Root directory for all test results created by the Bitrise CLI | required | `$BITRISE_TEST_RESULT_DIR` |
| `verbose_log` | Enable verbose logging? | required | `no` |
</details>

<details>
<summary>Outputs</summary>
There are no outputs defined in this step
</details>

## 🙋 Contributing

We welcome [pull requests](https://github.com/bitrise-steplib/step-custom-test-results-export/pulls) and [issues](https://github.com/bitrise-steplib/step-custom-test-results-export/issues) against this repository.

For pull requests, work on your changes in a forked repository and use the Bitrise CLI to [run step tests locally](https://docs.bitrise.io/en/bitrise-ci/bitrise-cli/running-your-first-local-build-with-the-cli.html).

Note: this step's end-to-end tests (defined in e2e/bitrise.yml) are working with secrets which are intentionally not stored in this repo. External contributors won't be able to run those tests. Don't worry, if you open a PR with your contribution, we will help with running tests and make sure that they pass.


Learn more about developing steps:

- [Create your own step](https://docs.bitrise.io/en/bitrise-ci/workflows-and-pipelines/developing-your-own-bitrise-step/developing-a-new-step.html)
