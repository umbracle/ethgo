import React from "react";
/**
 * SkipNavLink
 *
 * Renders a link that remains hidden until focused to skip to the main content.
 *
 * @see Docs https://reach.tech/skip-nav#skipnavlink
 */
export declare const SkipNavLink: React.FC<SkipNavLinkProps>;
/**
 * @see Docs https://reach.tech/skip-nav#skipnavlink-props
 */
export declare type SkipNavLinkProps = {
    /**
     * Allows you to change the text for your preferred phrase or localization.
     *
     * @see Docs https://reach.tech/skip-nav#skipnavlink-children
     */
    children?: React.ReactNode;
    /**
     * An alternative ID for `SkipNavContent`. If used, the same value must be
     * provided to the `id` prop in `SkipNavContent`.
     *
     * @see Docs https://reach.tech/skip-nav#skipnavlink-contentid
     */
    contentId?: string;
} & Omit<React.HTMLAttributes<HTMLAnchorElement>, "href">;
/**
 * SkipNavContent
 *
 * Renders a div as the target for the link.
 *
 * @see Docs https://reach.tech/skip-nav#skipnavcontent
 */
export declare const SkipNavContent: React.FC<SkipNavContentProps>;
/**
 * @see Docs https://reach.tech/skip-nav#skipnavcontent-props
 */
export declare type SkipNavContentProps = {
    /**
     * You can place the `SkipNavContent` element as a sibling to your main
     * content or as a wrapper.
     *
     * Keep in mind it renders a `div`, so it may mess with your CSS depending on
     * where it’s placed.
     *
     * @example
     *   <SkipNavContent />
     *   <YourMainContent />
     *   // vs.
     *   <SkipNavContent>
     *     <YourMainContent/>
     *   </SkipNavContent>
     *
     * @see Docs https://reach.tech/skip-nav#skipnavcontent-children
     */
    children?: React.ReactNode;
    /**
     * An alternative ID. If used, the same value must be provided to the
     * `contentId` prop in `SkipNavLink`.
     *
     * @see Docs https://reach.tech/skip-nav#skipnavcontent-id
     */
    id?: string;
} & Omit<React.HTMLAttributes<HTMLDivElement>, "id">;
