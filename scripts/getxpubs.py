import hashlib
from typing import NamedTuple

import base58
import click
from construct import Bytes, GreedyBytes, Int8ub, PascalString, Prefixed, Struct
from ledgerwallet.client import LedgerClient
from ledgerwallet.params import Bip32Path
from ledgerwallet.transport import enumerate_devices

BTCHIP_INS_GET_WALLET_PUBLIC_KEY = 0x40
BIP32_HARDEN = 0x80000000
BIP32_HD_VERSION_MAP = {
    "mainnet": 0x0488B21E,
    "testnet": 0x043587CF,
}
BIP32_HD_VERSION_MAP["regtest"] = BIP32_HD_VERSION_MAP["testnet"]

GetPubKeyResponse = Struct(
    public_key=Prefixed(Int8ub, GreedyBytes),
    address=PascalString(Int8ub, "utf-8"),
    chain_code=Bytes(32),
)


class ExtendedPublicKey(NamedTuple):
    version: int  # 4 bytes
    depth: int  # 1 byte
    parent_fingerprint: bytes  # 4 bytes
    child_num: int  # 4 bytes
    chaincode: bytes  # 32 bytes
    pubkey: bytes  # 32 bytes

    def serialize(self) -> str:
        version = self.version.to_bytes(length=4, byteorder="big")
        depth = self.depth.to_bytes(length=1, byteorder="big")
        child_num = self.child_num.to_bytes(length=4, byteorder="big")

        extended_key_bytes = (
            version
            + depth
            + self.parent_fingerprint
            + child_num
            + self.chaincode
            + self.pubkey
        )
        checksum = hash256(extended_key_bytes)[:4]
        return base58.b58encode(extended_key_bytes + checksum).decode()


def sha256(s) -> bytes:
    return hashlib.new("sha256", s).digest()


def ripemd160(s) -> bytes:
    return hashlib.new("ripemd160", s).digest()


def hash256(s) -> bytes:
    return sha256(sha256(s))


def hash160(s) -> bytes:
    return ripemd160(sha256(s))


def compress_public_key(public_key: bytes) -> bytes:
    # [TODO] - implement in ledgerwallet.crypto.ecc, or use ecdsa (lib)
    if len(public_key) == 64 + 1 and public_key[0] == 0x04:
        if public_key[64] & 1:
            return b"\x03" + public_key[1 : 32 + 1]
        else:
            return b"\x02" + public_key[1 : 32 + 1]
    elif len(public_key) == 32 + 1 and public_key[0] in (0x02, 0x03):
        return public_key
    else:
        raise ValueError("Invalid public key format")


def get_xpub_from_path(client: LedgerClient, path: str, network: str) -> str:
    # Get Pubkey and chaincode
    response = client.apdu_exchange(
        BTCHIP_INS_GET_WALLET_PUBLIC_KEY, Bip32Path.build(path)
    )
    r = GetPubKeyResponse.parse(response)
    pubkey = compress_public_key(r.public_key)
    chain_code = r.chain_code

    # Get Parent Path pubkey and compute its fingerprint:
    parent_path = "/".join(path.split("/")[:-1])
    response = client.apdu_exchange(
        BTCHIP_INS_GET_WALLET_PUBLIC_KEY, Bip32Path.build(parent_path)
    )
    r = GetPubKeyResponse.parse(response)
    parent_public_key = r.public_key

    extended_key = ExtendedPublicKey(
        version=BIP32_HD_VERSION_MAP[network],
        depth=len(path.split("/")),
        parent_fingerprint=hash160(compress_public_key(parent_public_key))[:4],
        child_num=BIP32_HARDEN,
        chaincode=chain_code,
        pubkey=pubkey,
    )
    return extended_key.serialize()


def get_client() -> LedgerClient:
    for device in enumerate_devices():
        return LedgerClient(device)
    else:
        raise ConnectionError("No Ledger device has been found.")


@click.command()
@click.option(
    "--scheme",
    type=click.Choice(["BIP44", "BIP49", "BIP84"], case_sensitive=False),
    required=True,
)
@click.option(
    "--network",
    type=click.Choice(["mainnet", "testnet", "regtest"], case_sensitive=False),
    required=True,
)
@click.option("--account", type=int, required=True)
def main(scheme, network, account):
    if scheme == "BIP44":
        path = f"44'/0'/{account}'"
        kind = "Legacy"
    elif scheme == "BIP49":
        path = f"49'/0'/{account}'"
        kind = "Segwit"
    elif scheme == "BIP84":
        path = f"84'/0'/{account}'"
        kind = "Native Segwit"
    else:
        raise ValueError(f"Bad derivation scheme: {scheme}")

    client = get_client()
    xpub = get_xpub_from_path(client, path, network)
    click.secho(f"{kind} ({scheme}) BTC {network} xPub: {xpub}", fg="green")


if __name__ == "__main__":
    main()
