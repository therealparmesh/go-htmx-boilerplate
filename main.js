import htmx from 'htmx.org';
import { Notyf } from 'notyf';

const me = () => window.document.currentScript?.parentElement;
window.htmx = htmx;
window.me = me;

document.addEventListener('DOMContentLoaded', () => {
  const notyf = new Notyf({ position: { x: 'right', y: 'top' } });

  document.body.addEventListener('successNotification', (event) => {
    notyf.success(event.detail.value);
  });

  document.body.addEventListener('errorNotification', (event) => {
    notyf.error(event.detail.value);
  });
});
