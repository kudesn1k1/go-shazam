import { describe, it, expect, afterEach } from 'vitest';
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils';
import AuthModal from './AuthModal.vue';

// AuthModal uses `<Teleport to="body">`. wrapper.find does not always reach
// teleported content in happy-dom, so we query document.body directly and
// fire native input/click events. The wrapper is still used to read emitted
// events and to setProps.

describe('AuthModal', () => {
  const wrappers: VueWrapper[] = [];

  afterEach(() => {
    while (wrappers.length) wrappers.pop()!.unmount();
    document.body.innerHTML = '';
  });

  function mountModal(props: { modelValue: boolean; mode: 'login' | 'register' }) {
    const w = mount(AuthModal, { props, attachTo: document.body });
    wrappers.push(w);
    return w;
  }

  function setInput(selector: string, value: string) {
    const el = document.querySelector(selector) as HTMLInputElement;
    el.value = value;
    el.dispatchEvent(new Event('input', { bubbles: true }));
  }

  it('emits submit with email + password on valid login form', async () => {
    const wrapper = mountModal({ modelValue: true, mode: 'login' });

    setInput('#email', 'user@example.com');
    setInput('#password', 'any-password');
    await flushPromises();

    const form = document.querySelector('form') as HTMLFormElement;
    form.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
    await flushPromises();

    const submitEvents = wrapper.emitted('submit');
    expect(submitEvents).toBeTruthy();
    expect(submitEvents![0]![0]).toEqual({
      mode: 'login',
      email: 'user@example.com',
      password: 'any-password',
    });
  });

  it('shows validation error when password too short in register mode', async () => {
    mountModal({ modelValue: true, mode: 'register' });

    setInput('#email', 'user@example.com');
    setInput('#password', 'short');
    setInput('#confirm', 'short');
    await flushPromises();

    const form = document.querySelector('form') as HTMLFormElement;
    form.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
    await flushPromises();

    expect(document.body.textContent).toContain('Password must be at least 8 characters');
  });

  it('shows validation error when register confirm does not match', async () => {
    mountModal({ modelValue: true, mode: 'register' });

    setInput('#email', 'user@example.com');
    setInput('#password', 'long-enough-pw');
    setInput('#confirm', 'different-pw-here');
    await flushPromises();

    const form = document.querySelector('form') as HTMLFormElement;
    form.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
    await flushPromises();

    expect(document.body.textContent).toContain('Passwords do not match');
  });

  it('emits update:modelValue=false when close button clicked', async () => {
    const wrapper = mountModal({ modelValue: true, mode: 'login' });

    const closeBtn = document.querySelector('.close-btn') as HTMLButtonElement;
    closeBtn.click();
    await flushPromises();

    expect(wrapper.emitted('update:modelValue')).toBeTruthy();
    expect(wrapper.emitted('update:modelValue')![0]).toEqual([false]);
  });

  it('switching mode resets form fields', async () => {
    const wrapper = mountModal({ modelValue: true, mode: 'login' });

    setInput('#email', 'keep-me@example.com');
    await flushPromises();

    await wrapper.setProps({ mode: 'register' });
    await flushPromises();

    const email = document.querySelector('#email') as HTMLInputElement;
    expect(email.value).toBe('');
  });
});
