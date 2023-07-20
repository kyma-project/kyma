## How to locally run and debug Python Function nad runtime

* Create [venv](https://docs.python.org/3/library/venv.html)
* Active venv:
  ```bash
  source venv/bin/activete
  ```
* Create the `function` directory with `handler.py` and `requirements.txt`
* Install dependencies from runtime and function:
    ```bash
    pip install -r requirements.txt
    pip install -r kubeless/requirements.txt
    ```
* Set following envs:
    ```bash
    export FUNCTION_PATH=./function
    export MOD_NAME=handler
    export FUNC_HANDLER=main
    ```
* Run function from the terminal.
    ```bash
    python3 kubeless.py
    ```

