## How to run/debug Nodejs function and runtime locally

* Copy `package.json` from desired nodejs version * 
* Create directory `function` with `handler.js` and `package.json`
* Install dependencies from runtime and function:
```bash
npm install
npm install function/
```
* Set proper environments variables TODO: check if it's needed
```bash
export MOD_NAME=handler
export FUNC_HANDLER=main
```
* Run function from terminal.
```bash
npm start
```

