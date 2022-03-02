import { ReactNode } from 'react';
interface InnerText {
    (jsx: ReactNode): string;
    default: InnerText;
}
declare const innerText: InnerText;
export = innerText;
