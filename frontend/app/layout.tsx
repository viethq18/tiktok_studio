import type { Metadata } from 'next';
import { cookies } from 'next/headers';
import { Be_Vietnam_Pro, Inter, Montserrat, Nunito, Playfair_Display, Roboto } from 'next/font/google';

import { Providers } from './providers';
import { LOCALE_COOKIE, parseLocale } from '@/lib/i18n/locale';
import './globals.css';

// These families mirror backend/internal/fontkit — the editor must be able to
// render every font the design generator is allowed to choose (§40).
const inter = Inter({ subsets: ['latin', 'vietnamese'], weight: ['400', '600', '700', '800'], variable: '--font-inter' });
const roboto = Roboto({ subsets: ['latin', 'vietnamese'], weight: ['400', '500', '700', '900'], variable: '--font-roboto' });
const montserrat = Montserrat({ subsets: ['latin', 'vietnamese'], weight: ['400', '600', '700', '800'], variable: '--font-montserrat' });
const nunito = Nunito({ subsets: ['latin', 'vietnamese'], weight: ['400', '600', '700'], variable: '--font-nunito' });
const playfair = Playfair_Display({ subsets: ['latin', 'vietnamese'], weight: ['400', '700', '900'], variable: '--font-playfair' });
const beVietnam = Be_Vietnam_Pro({ subsets: ['latin', 'vietnamese'], weight: ['400', '600', '700', '800'], variable: '--font-be-vietnam' });


export const metadata: Metadata = {
  title: 'TikTok Carousel Studio',
  description:
    'Describe your channel in one sentence and get a finished TikTok carousel: research, copy, layout, photography and caption.',
};

export default async function RootLayout({ children }: { children: React.ReactNode }) {
  // Resolving the locale on the server means an English visitor's first paint is
  // already English — no flash of the default language after hydration.
  const locale = parseLocale((await cookies()).get(LOCALE_COOKIE)?.value);

  return (
    <html lang={locale}>
      <body
        className={`${inter.variable} ${roboto.variable} ${montserrat.variable} ${nunito.variable} ${playfair.variable} ${beVietnam.variable} bg-neutral-50 font-sans text-neutral-900 antialiased`}
      >
        <Providers locale={locale}>{children}</Providers>
      </body>
    </html>
  );
}
