
import Link from 'next/link'

export default function EIPLink({children, num}) {
    return <Link href={`https://github.com/ethereum/EIPs/blob/master/EIPS/eip-${num}.md`}>{children}</Link>
}
