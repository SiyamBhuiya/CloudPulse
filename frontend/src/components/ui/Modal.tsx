import { useEffect, useRef, type ReactNode } from "react";

interface Props {
  open: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
}

export default function Modal({ open, onClose, title, children }: Props) {
  const ref = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    const d = ref.current;
    if (!d) return;
    if (open && !d.open) d.showModal();
    if (!open && d.open) d.close();
  }, [open]);

  return (
    <dialog
      ref={ref}
      onClose={onClose}
      onClick={(e) => e.target === ref.current && onClose()} // click on backdrop
      aria-labelledby="modal-title"
      className="m-auto w-[min(460px,calc(100%-32px))] rounded-2xl border border-line bg-panel p-6 text-ink backdrop:bg-black/60"
    >
      <h2 id="modal-title" className="mb-1 text-xl font-bold">
        {title}
      </h2>
      {open && children}
    </dialog>
  );
}