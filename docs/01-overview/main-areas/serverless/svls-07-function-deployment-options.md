---
title: Best Practices for Function Development
---


# Overview - Its all about Custom Resources

Kyma serverless introduces a [`Function`](https://kyma-project.io/docs/kyma/latest/05-technical-reference/00-custom-resources/svls-01-function/) Custom Resource Definition (CRD) as an extension to the kubernetes API server.
Defining a Function in kyma essentially means creating a new instance of Function Custom Resource (CR). However, the content of Function CR specification may become quite long. It consists of the code (or git reference to the code), dependencies, runtime specification, build-time specification,  etc. Additionally there are other CRs that are relevant for function developer - I.e `APIRule` ( defining how function is exposed to the outside world), `Subscription` (defining which cloud events should trigger the function) and others.
All of that makes it cumbersome to define the setup by hand. 
The following sections will guide you through the best practices for function development. You will find hints that will be helpful for you at any stage of your development journey.

# Use UI to explore

At the begining of your kyma journey you will probably want to evaluate serverless and draft a few functions.
Kyma Dashboard is perfect to gain basic experience and start the journey with kyma functions. Dashboard consists of UI components dedicated for serverless that will help you draft your first functions by putting the code directly in the browser via Web IDE.
Kyma Dashboard will also help you expose your function via HTTP, define environment variables, subscribe it to cloud events, bind service instances and even present you function logs - all in single place.

Get started with [Function UI](https://kyma-project.io/docs/kyma/latest/03-tutorials/00-serverless/svls-01-create-inline-function/)

![function-ui](./assets/function-ui.png)

# Use Kyma CLI for better development experience

Defining function from the UI is very quick and easy but most likely this approach would be not enough to satisfy your needs as a developer. Take for example the ability to code your function from your favourite IDE or being able to run and debug the function on your local machine before actually deploying in the kyma runtime - Those aspects may be desired especially if you want to code and test a bit more complex ones. Also, you would probably want to avoid recreating same functions manually from the UI on a differnt environment. In the end having deployable artefacts is more desirable. This is where kyma CLI comes in handy as it allows you to keep you function's code and configuration in a form of a workspace. 

Initialise a scaffold for a brand new function via `kyma init function` command or fetch the current state of an existing function deployed in your kyma runtime via `kyma sync function`.
Focus on the function code and develop it from you favourite IDE. Configure your functions directly in the [`config.yaml` manifest file](https://kyma-project.io/docs/kyma/latest/05-technical-reference/svls-06-function-configuration-file/)
>>NOTE: Use `kyma init function --vscode` to generate json schema which can be used in VScode for autocompletion.
>>

Kyma CLI helps you run your code locally with a single `kyma run function` command. It runs the function using your local docker deamon with the same runtime docker context as if it was run in the kyma runtime. 
>>NOTE: Use `kyma run function` with `--hot-deploy` and spare yourself unnecessary restarts of the functions whenever you test a changed function logic. Also, use [`--debug` option](https://kyma-project.io/docs/kyma/latest/03-tutorials/00-serverless/svls-05-debug-function) to allow connecting with your favourite debugger.
>>

![kyma-cli-functions](./assets/kyma-cli-functions.png)

Having written and tested your function locally, simply deploy it to the kyma runtime with `kyma apply function` command used in the folder of your function's workspace. It will read the files, translate it to kubernetes manifests and deploy the function.


# Deploy via CI/CD

Kyma UI helps you getting started. Kyma CLI helps you iterate and develop functions. 
But at the end of the day you probably want an automated deployment of your functions. Or rather, of your whole application, where functions are just part of it.
It all comes down to being able to deploy k8s applications on different kyma runtimes in a GitOps fashion and, for the sake of simplicity, the deployment approach for functions should not differ from deployment of other kubernetes workloads, config-maps or secrets.

So in the end what you need is those yaml manifests for everything - including functions.

Fortunatelly, Kyma CLI helps you generate the yaml mainfests matching your `config.yaml` file you crafted in earlier phase.
Use `--dry-run` option of the `kyma apply function` command to generate kubernetes manifests that will include the function CR itself but also all the related CRs (i.e ApiRules, Subscriptions, etc).

```bash
kyma apply function --dry-run --ci -o yaml > my-function.yaml
```  

The generated manifest should be part of all the manifests that define your application and pushed to the git repository.
Deploy everything in a consistent way either via CI/CD or via GitOps operators ( i.e [fluxcd](https://kyma-project.io/docs/kyma/latest/03-tutorials/00-serverless/svls-06-sync-function-with-gitops/) ) installed in your kyma runtime.

>>NOTE: Source Function Code  directly from Git Repository

Kyma Functions come in two types : `git` and `inline`.
[Git type](https://kyma-project.io/docs/kyma/latest/03-tutorials/00-serverless/svls-02-create-git-function/) allows you to configure a git repository as a source of your function code instead of creating it "inline".
This allows you to skip rendering of k8s manifests and deploying them everytime you made a change in the function code or dependencies. Simply push the changes to the referenced got repository and the serverless controller will rebuild the functions that is deployed in your kyma runtime. 

Please have a look at the following example that illustrates how you can setup your git project
<!-- KK TODO. Link example -->


>>

