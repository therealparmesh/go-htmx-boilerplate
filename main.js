import htmx from 'htmx.org';
import { Notyf } from 'notyf';

window.htmx = htmx;

window.me = (previous) =>
  previous
    ? window.document.currentScript?.previousElementSibling
    : window.document.currentScript?.parentElement;

document.addEventListener('DOMContentLoaded', () => {
  const notyf = new Notyf({ position: { x: 'right', y: 'top' } });

  document.body.addEventListener('successNotification', (event) => {
    notyf.success(event.detail.value);
  });

  document.body.addEventListener('errorNotification', (event) => {
    notyf.error(event.detail.value);
  });
});
