## How to run/debug Python function and runtime locally

* Create venv if not creaded
* Active venv
* Create directory `kubeless` with `handler.py` and `requirements`
* Install dependencies from runtime and function:
```bash
pip install -r requirements.txt
pip install -r kubeless/requirements.txt
```
* Set following envs:
```bash
export PYTHON_PATH=kubeless
export MOD_
export HANDLER=
```
* Run function from terminal.
```bash
npm start
```

