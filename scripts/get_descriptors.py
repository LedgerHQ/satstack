import hashlib
from dataclasses import dataclass, field
from typing import List, Literal

import base58
import click
from construct import Bytes, GreedyBytes, Int8ub, PascalString, Prefixed, Struct
from ledgerwallet.client import LedgerClient
from ledgerwallet.params import Bip32Path as bip32_path
from ledgerwallet.transport import enumerate_devices

BTCHIP_INS_GET_WALLET_PUBLIC_KEY = 0x40
BIP32_HD_VERSION_MAINNET = 0x0488B21E
SCHEME = Literal["legacy", "segwit", "native_segwit"]

GetPubKey = Struct(
    public_key=Prefixed(Int8ub, GreedyBytes),
    address=PascalString(Int8ub, "utf-8"),
    chain_code=Bytes(32),
)


@dataclass
class Derivation:
    _path_list: List["Level"] = field(default_factory=list)

    def __truediv__(self, level: "Level") -> "Derivation":
        return Derivation(self._path_list + [level])

    @property
    def parent(self) -> "Derivation":
        return Derivation(self._path_list[:-1])

    @property
    def path(self) -> str:
        return "/".join(str(level) for level in self._path_list)

    @property
    def depth(self) -> int:
        return len(self._path_list)


@dataclass
class Level:
    _value: int

    BIP32_HARDEN_BIT = 0x80000000

    def harden(self) -> "Level":
        return Level(self._value + self.BIP32_HARDEN_BIT)

    h = harden

    def __str__(self) -> str:
        if self._value & self.BIP32_HARDEN_BIT:
            value = self._value - self.BIP32_HARDEN_BIT
            return f"{value}'"
        return f"{self._value}"


@dataclass
class ExtendedPublicKey:
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

    def to_descriptor(
        self, scheme: SCHEME, derivation: Derivation, change: bool
    ) -> str:
        key_origin = f"{self.parent_fingerprint.hex()}/{derivation.path}"

        change_index = 1 if change else 0
        fragment = f"[{key_origin}]{self.serialize()}/{change_index}/*"

        if scheme == "legacy":
            return f"pkh({fragment})"
        elif scheme == "segwit":
            return f"sh(wpkh({fragment}))"
        elif scheme == "native_segwit":
            return f"wpkh({fragment})"

        raise ValueError(f"Invalid scheme: {scheme}")


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


def get_pubkey_from_path(client: LedgerClient, derivation: Derivation):
    response = client.apdu_exchange(
        BTCHIP_INS_GET_WALLET_PUBLIC_KEY, bip32_path.build(derivation.path)
    )
    r = GetPubKey.parse(response)
    pubkey = compress_public_key(r.public_key)
    chain_code = r.chain_code
    return pubkey, chain_code


def derive_extended_public_key(
    client: LedgerClient, derivation: Derivation
) -> ExtendedPublicKey:
    pubkey, chain_code = get_pubkey_from_path(client, derivation)
    parent_pubkey, _ = get_pubkey_from_path(client, derivation.parent)

    return ExtendedPublicKey(
        version=BIP32_HD_VERSION_MAINNET,
        depth=derivation.depth,
        parent_fingerprint=hash160(parent_pubkey)[:4],
        child_num=0x80000000,
        chaincode=chain_code,
        pubkey=pubkey,
    )


def get_client() -> LedgerClient:
    for device in enumerate_devices():
        return LedgerClient(device)
    raise ConnectionError("No Ledger device has been found.")


def get_derivation_from_scheme(scheme: SCHEME, account: int):
    m = Derivation()

    if scheme == "legacy":
        return m / Level(44).h() / Level(0).h() / Level(account).h()
    elif scheme == "segwit":
        return m / Level(49).h() / Level(0).h() / Level(account).h()
    elif scheme == "native_segwit":
        return m / Level(84).h() / Level(0).h() / Level(account).h()

    raise ValueError(f"Bad derivation scheme: {scheme}")


def derive_output_descriptors(client: LedgerClient, scheme: SCHEME, account: int):
    derivation = get_derivation_from_scheme(scheme, account)
    extended_key = derive_extended_public_key(client, derivation)

    yield extended_key.to_descriptor(scheme=scheme, derivation=derivation, change=False)

    yield extended_key.to_descriptor(scheme=scheme, derivation=derivation, change=True)


@click.command()
@click.option(
    "--scheme", type=click.Choice(["legacy", "segwit", "native_segwit"]), required=True,
)
@click.option("--account", type=int, required=True)
def main(scheme: SCHEME, account):
    result = ()
    client = get_client()
    for descriptor in derive_output_descriptors(client, scheme, account):
        result += ({"descriptor": descriptor},)

    click.secho(
        f"Bitcoin output descriptors: scheme={scheme} network=mainnet", fg="green",
    )
    for each in result:
        click.echo(f"\t{each['descriptor']}")


if __name__ == "__main__":
    main()
