import { useHead } from '@unhead/vue';

// Mark a route as private — keeps it out of search indices and tab title noise.
// Use on every authenticated/admin page.
export function useNoindex(title: string) {
  useHead({
    title: `${title} — Go Shazam`,
    meta: [
      { name: 'robots', content: 'noindex, nofollow' },
    ],
  });
}
