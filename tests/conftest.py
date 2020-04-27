import os
import subprocess
import time

import pytest
import requests


@pytest.fixture(scope="session")
def server():
    print("Building project")
    subprocess.Popen(
        ["make", "build"],
        cwd=os.path.join(os.getcwd(), ".."),
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
    ).wait()

    print("Launch live server")
    process = subprocess.Popen(
        ["./main"],
        cwd=os.path.join(os.getcwd(), ".."),
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
    )

    print("Wait for server to accept connections")
    time.sleep(10)

    while True:
        health = requests.get("http://localhost:8080/blockchain/v3/explorer/_health")
        if health.json() == {"Status": "OK"}:
            break
        time.sleep(1)

    try:
        yield
    finally:
        process.terminate()
        os.remove(os.path.join(os.getcwd(), "..", "main"))
