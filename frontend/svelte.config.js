import adapter from '@sveltejs/adapter-auto';
import preprocess from 'svelte-preprocess';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: preprocess({
    typescript: true
  }),
  kit: {
    adapter: adapter(),
    alias: {
      $lib: 'src/lib',
      $components: 'src/lib/components',
      $stores: 'src/lib/stores',
      $api: 'src/lib/api',
      $types: 'src/lib/types',
      $helpers: 'src/lib/helpers'
    }
  }
};

export default config;
