import requests


def comparator(route):
    remote_endpoint = f"http://bitcoin-mainnet.explorers.prod.aws.ledger.fr/{route}"
    local_endpoint = f"http://localhost:8080/{route}"

    remote_response = requests.get(remote_endpoint)
    local_response = requests.get(local_endpoint)

    assert local_response.status_code == remote_response.status_code
    assert local_response.text == remote_response.text


def test_health(server):
    comparator("blockchain/v3/explorer/_health")


def test_get_current_block(server):
    comparator("blockchain/v3/blocks/current")
