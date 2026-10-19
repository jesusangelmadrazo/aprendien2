import { useEffect, useState } from "react";
import {
  createBlock,
  deleteBlock,
  listMyBlocks,
  setOverride,
  listMyOverrides,
  deleteOverride,
} from "../../api/availability";
import type { AvailabilityBlock, ScheduleOverride } from "../../types";

const DAY_NAMES = ["Domingo", "Lunes", "Martes", "Miércoles", "Jueves", "Viernes", "Sábado"];
const DAY_SHORT = ["Dom", "Lun", "Mar", "Mié", "Jue", "Vie", "Sáb"];
const MONTHS = ["ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sep", "oct", "nov", "dic"];

interface TimeValue {
  h: number;
  m: number;
}

function pad(n: number) {
  return String(n).padStart(2, "0");
}

function fmtTime(t: TimeValue) {
  return `${pad(t.h)}:${pad(t.m)}`;
}

function todayISO() {
  const d = new Date();
  const offset = d.getTimezoneOffset();
  return new Date(d.getTime() - offset * 60000).toISOString().slice(0, 10);
}

function TimeSelect({ value, onChange }: { value: TimeValue; onChange: (v: TimeValue) => void }) {
  return (
    <div className="time-select">
      <select value={value.h} onChange={(e) => onChange({ ...value, h: Number(e.target.value) })}>
        {Array.from({ length: 24 }, (_, h) => (
          <option key={h} value={h}>
            {pad(h)}
          </option>
        ))}
      </select>
      <span>:</span>
      <select value={value.m} onChange={(e) => onChange({ ...value, m: Number(e.target.value) })}>
        {[0, 15, 30, 45].map((m) => (
          <option key={m} value={m}>
            {pad(m)}
          </option>
        ))}
      </select>
    </div>
  );
}

export default function AvailabilityEditor() {
  const [blocks, setBlocks] = useState<AvailabilityBlock[]>([]);
  const [overrides, setOverrides] = useState<ScheduleOverride[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [openAddFor, setOpenAddFor] = useState<number | null>(null);
  const [draftStart, setDraftStart] = useState<TimeValue>({ h: 9, m: 0 });
  const [draftEnd, setDraftEnd] = useState<TimeValue>({ h: 12, m: 0 });

  const [exceptionFormOpen, setExceptionFormOpen] = useState(false);
  const [excDate, setExcDate] = useState(todayISO());
  const [excType, setExcType] = useState<"off" | "custom">("off");
  const [excStart, setExcStart] = useState<TimeValue>({ h: 9, m: 0 });
  const [excEnd, setExcEnd] = useState<TimeValue>({ h: 12, m: 0 });

  function refreshBlocks() {
    listMyBlocks().then(setBlocks);
  }

  function refreshOverrides() {
    listMyOverrides().then(setOverrides);
  }

  useEffect(() => {
    Promise.all([listMyBlocks(), listMyOverrides()])
      .then(([b, o]) => {
        setBlocks(b);
        setOverrides(o);
      })
      .finally(() => setLoading(false));
  }, []);

  async function handleAddBlock(day: number) {
    setError(null);
    const start = fmtTime(draftStart);
    const end = fmtTime(draftEnd);
    if (start >= end) {
      setError("La hora de inicio debe ser antes que la de fin");
      return;
    }
    try {
      await createBlock({ day_of_week: day, start_time: start, end_time: end });
      setOpenAddFor(null);
      refreshBlocks();
    } catch (err: any) {
      setError(err.response?.data?.error ?? "No se pudo agregar el horario");
    }
  }

  async function handleRemoveBlock(id: string) {
    await deleteBlock(id);
    refreshBlocks();
  }

  async function handleSaveException() {
    setError(null);
    try {
      if (excType === "off") {
        await setOverride({ date: excDate, is_day_off: true });
      } else {
        const start = fmtTime(excStart);
        const end = fmtTime(excEnd);
        if (start >= end) {
          setError("La hora de inicio debe ser antes que la de fin");
          return;
        }
        await setOverride({ date: excDate, is_day_off: false, start_time: start, end_time: end });
      }
      setExceptionFormOpen(false);
      refreshOverrides();
    } catch (err: any) {
      setError(err.response?.data?.error ?? "No se pudo guardar la excepción");
    }
  }

  async function handleRemoveException(date: string) {
    await deleteOverride(date.slice(0, 10));
    refreshOverrides();
  }

  if (loading) return <p>Cargando...</p>;

  return (
    <section className="availability-editor">
      <h2>Mi agenda semanal</h2>
      <p className="hint">
        Se repite todas las semanas. Tocá "+ Agregar" en cualquier día para sumar un horario — así ves de un vistazo
        dónde tenés huecos.
      </p>

      <div className="week-card">
        {DAY_NAMES.map((name, day) => {
          const dayBlocks = blocks.filter((b) => b.DayOfWeek === day);
          return (
            <div key={day} className="day-row">
              <div className="day-label">{name}</div>
              <div className="day-body">
                {dayBlocks.length === 0 && openAddFor !== day && <span className="empty-hint">Sin horarios</span>}
                {dayBlocks.map((b) => (
                  <span key={b.ID} className="block-chip">
                    {b.StartTime} – {b.EndTime}
                    <button aria-label="Quitar horario" onClick={() => handleRemoveBlock(b.ID)}>
                      ×
                    </button>
                  </span>
                ))}
                {openAddFor === day ? (
                  <div className="add-form">
                    <TimeSelect value={draftStart} onChange={setDraftStart} />
                    <span className="add-form-sep">a</span>
                    <TimeSelect value={draftEnd} onChange={setDraftEnd} />
                    <div className="add-form-actions">
                      <button className="icon-btn confirm" aria-label="Confirmar" onClick={() => handleAddBlock(day)}>
                        ✓
                      </button>
                      <button className="icon-btn cancel" aria-label="Cancelar" onClick={() => setOpenAddFor(null)}>
                        ×
                      </button>
                    </div>
                  </div>
                ) : (
                  <button
                    className="add-btn"
                    onClick={() => {
                      setDraftStart({ h: 9, m: 0 });
                      setDraftEnd({ h: 12, m: 0 });
                      setOpenAddFor(day);
                    }}
                  >
                    + Agregar
                  </button>
                )}
              </div>
            </div>
          );
        })}
      </div>

      {error && <p className="form-error">{error}</p>}

      <div className="section-header">
        <h3>Excepciones puntuales</h3>
        <p className="hint">
          Para una fecha específica que no sigue tu horario de siempre — un día libre, o un horario distinto solo
          para esa semana. Mientras exista, pisa tu agenda semanal para esa fecha.
        </p>
      </div>

      <div className="exceptions-card">
        {overrides.length === 0 ? (
          <p className="exceptions-empty">No tenés excepciones cargadas.</p>
        ) : (
          overrides.map((o) => {
            const d = new Date(o.Date);
            return (
              <div key={o.ID} className="exception-row">
                <div className="exception-date">
                  <span className="month">{MONTHS[d.getUTCMonth()]}</span>
                  {d.getUTCDate()}
                </div>
                <div className="exception-info">
                  <div className={"what" + (o.IsDayOff ? " day-off" : "")}>
                    {o.IsDayOff ? "Día libre" : `${o.StartTime} – ${o.EndTime}`}
                  </div>
                  <div className="weekday">{DAY_SHORT[d.getUTCDay()]}</div>
                </div>
                <button className="add-btn" onClick={() => handleRemoveException(o.Date)}>
                  Quitar
                </button>
              </div>
            );
          })
        )}
      </div>

      {exceptionFormOpen ? (
        <div className="exception-form">
          <label>
            <span className="field-label">Fecha</span>
            <input type="date" min={todayISO()} value={excDate} onChange={(e) => setExcDate(e.target.value)} />
          </label>

          <div className="mode-toggle">
            <button className={excType === "off" ? "active" : ""} onClick={() => setExcType("off")}>
              Día libre
            </button>
            <button className={excType === "custom" ? "active" : ""} onClick={() => setExcType("custom")}>
              Horario especial
            </button>
          </div>

          {excType === "custom" && (
            <div className="add-form">
              <TimeSelect value={excStart} onChange={(v) => setExcStart(v)} />
              <span className="add-form-sep">a</span>
              <TimeSelect value={excEnd} onChange={(v) => setExcEnd(v)} />
            </div>
          )}

          <div className="exception-form-actions">
            <button className="btn-ghost" onClick={() => setExceptionFormOpen(false)}>
              Cancelar
            </button>
            <button className="btn-primary" onClick={handleSaveException}>
              Guardar excepción
            </button>
          </div>
        </div>
      ) : (
        <button
          className="add-exception-btn"
          onClick={() => {
            setExcDate(todayISO());
            setExcType("off");
            setExceptionFormOpen(true);
          }}
        >
          + Agregar excepción
        </button>
      )}
    </section>
  );
}
