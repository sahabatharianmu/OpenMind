import { useTranslation } from 'react-i18next';
import { useCallback } from 'react';

/**
 * Provides locale-aware formatting for currency, dates, and numbers.
 * Uses the current i18n language to determine locale (en → en-US, id → id-ID).
 */
export function useFormatters() {
  const { i18n } = useTranslation();

  const locale = i18n.language === 'id' ? 'id-ID' : 'en-US';

  const formatCurrency = useCallback(
    (cents: number, currency = 'USD') => {
      return new Intl.NumberFormat(locale, {
        style: 'currency',
        currency,
      }).format(cents / 100);
    },
    [locale]
  );

  const formatCurrencyRaw = useCallback(
    (amount: number, currency = 'IDR') => {
      return new Intl.NumberFormat(locale, {
        style: 'currency',
        currency,
      }).format(amount);
    },
    [locale]
  );

  const formatDate = useCallback(
    (dateStr: string, style: 'short' | 'medium' | 'long' = 'medium') => {
      const options: Intl.DateTimeFormatOptions =
        style === 'short'
          ? { month: 'numeric', day: 'numeric', year: '2-digit' }
          : style === 'long'
          ? { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' }
          : { month: 'short', day: 'numeric', year: 'numeric' };
      return new Date(dateStr).toLocaleDateString(locale, options);
    },
    [locale]
  );

  const formatNumber = useCallback(
    (n: number) => new Intl.NumberFormat(locale).format(n),
    [locale]
  );

  return { formatCurrency, formatCurrencyRaw, formatDate, formatNumber, locale };
}
