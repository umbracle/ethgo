
import Link from 'next/link'
import GodocLink from "./godoc"

function Address({text='Address'}) {
    return <GodocLink href='#Address'>{`(${text})`}</GodocLink>
}

function Hash({text='Hash'}) {
    return <GodocLink href='#Hash'>{`(${text})`}</GodocLink>
}

function Block() {
    return <GodocLink href='#Block'>{'(Block)'}</GodocLink>
}

function Blocktag() {
    return <Link href={'/jsonrpc#block-tag'}>{'(BlockTag)'}</Link>
}

function ABI() {
    return <Link href={'/abi'}>{'(ABI)'}</Link>
}

function Transaction() {
    return <GodocLink href='#Transaction'>{'(Transaction)'}</GodocLink>
}

function Receipt() {
    return <GodocLink href='#Receipt'>{'(Receipt)'}</GodocLink>
}

export {
    Address,
    Hash,
    Block,
    Blocktag,
    Transaction,
    Receipt,
    ABI,
}
