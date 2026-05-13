import { describe, it, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import Pagination from './Pagination.vue';

describe('Pagination', () => {
  it('renders nothing when total fits on one page', () => {
    const wrapper = mount(Pagination, {
      props: { modelValue: 1, total: 10, limit: 20 },
    });
    expect(wrapper.find('.pagination').exists()).toBe(false);
  });

  it('emits update:modelValue when next button is clicked', async () => {
    const wrapper = mount(Pagination, {
      props: { modelValue: 2, total: 100, limit: 20 }, // 5 pages
    });

    const buttons = wrapper.findAll('button.page-btn');
    // First is prev "‹", then numbered pages, last is next "›"
    const nextBtn = buttons[buttons.length - 1]!;
    await nextBtn.trigger('click');

    expect(wrapper.emitted('update:modelValue')).toBeTruthy();
    expect(wrapper.emitted('update:modelValue')![0]).toEqual([3]);
  });

  it('emits update:modelValue when prev button is clicked', async () => {
    const wrapper = mount(Pagination, {
      props: { modelValue: 3, total: 100, limit: 20 },
    });

    const prevBtn = wrapper.findAll('button.page-btn')[0]!;
    await prevBtn.trigger('click');

    expect(wrapper.emitted('update:modelValue')![0]).toEqual([2]);
  });

  it('disables prev on page 1', () => {
    const wrapper = mount(Pagination, {
      props: { modelValue: 1, total: 100, limit: 20 },
    });
    const prevBtn = wrapper.findAll('button.page-btn')[0]!;
    expect(prevBtn.attributes('disabled')).toBeDefined();
  });

  it('disables next on last page', () => {
    const wrapper = mount(Pagination, {
      props: { modelValue: 5, total: 100, limit: 20 },
    });
    const buttons = wrapper.findAll('button.page-btn');
    const nextBtn = buttons[buttons.length - 1]!;
    expect(nextBtn.attributes('disabled')).toBeDefined();
  });

  it('marks the current page button as active', () => {
    const wrapper = mount(Pagination, {
      props: { modelValue: 3, total: 100, limit: 20 },
    });
    const active = wrapper.find('button.page-btn.active');
    expect(active.exists()).toBe(true);
    expect(active.text()).toBe('3');
  });

  it('emits direct page number when a numbered button is clicked', async () => {
    const wrapper = mount(Pagination, {
      props: { modelValue: 3, total: 100, limit: 20 },
    });
    // visible pages = [1,2,3,4,5]. Find button labeled "5".
    const fiveBtn = wrapper.findAll('button.page-btn').find((b) => b.text() === '5')!;
    await fiveBtn.trigger('click');
    expect(wrapper.emitted('update:modelValue')![0]).toEqual([5]);
  });
});
