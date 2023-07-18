## How to run/debug Nodejs function and runtime locally

1. Copy `package.json` from desired nodejs version
2. Create directory `function` with `handler.js` and `package.json`
3. Install dependencies from runtime and function:
```bash
npm install
npm install function/
```
4. Set proper environments variables TODO: check if it's needed
```bash
export MOD_NAME=handler
export FUNC_HANDLER=main
```
5. Run function from terminal.
```bash
npm start
```
