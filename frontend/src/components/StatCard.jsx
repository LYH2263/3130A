export function StatCard({ title, value, hint }) {
  return (
    <div className="rounded-2xl border border-slate-200 bg-base-100 p-5 shadow-card transition duration-300 hover:-translate-y-1 hover:shadow-xl">
      <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">{title}</p>
      <p className="mt-3 text-3xl font-bold text-slate-800">{value}</p>
      {hint ? <p className="mt-2 text-xs text-slate-500">{hint}</p> : null}
    </div>
  );
}
