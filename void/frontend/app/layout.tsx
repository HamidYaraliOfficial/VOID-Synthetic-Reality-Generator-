import type { Metadata } from 'next';
import './globals.css';
import { ThemeProvider } from '@/components/theme-provider';
import { I18nProvider } from '@/lib/i18n';
import { Sidebar } from '@/components/sidebar';

export const metadata: Metadata = {
  title: 'VOID — Synthetic Reality Generator',
  description: 'Create, simulate, control and analyze synthetic realities at massive scale.',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" dir="ltr" data-theme="windows" suppressHydrationWarning>
      <body className="font-sans">
        <ThemeProvider>
          <I18nProvider>
            <div className="flex min-h-screen">
              <Sidebar />
              <main className="flex-1 p-8 max-w-[1600px]">{children}</main>
            </div>
          </I18nProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
