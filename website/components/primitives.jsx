
import Link from 'next/link'

function Address({text='Address'}) {
    return <Link href={`/#address`}>{`(${text})`}</Link>
}

function Hash({text='Hash'}) {
    return <Link href={`/#hash`}>{`(${text})`}</Link>
}

function Block() {
    return <Link href={'/#block'}>{'(Block)'}</Link>
}

function Blocktag() {
    return <Link href={'/jsonrpc#block-tag'}>{'(BlockTag)'}</Link>
}

function Transaction() {
    return <Link href={'/#transaction'}>{'(Transaction)'}</Link>
}

function Receipt() {
    return <Link href={'/#receipt'}>{'(Receipt)'}</Link>
}

export {
    Address,
    Hash,
    Block,
    Blocktag,
    Transaction,
    Receipt,
}
