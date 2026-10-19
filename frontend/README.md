# Clase Drive — Frontend

React + Vite + TypeScript. Consume la API del backend en Go (`../clase-drive-backend`).

Plan completo del proyecto: `/Users/jesus.madrazo/.claude/plans/cheerful-enchanting-codd.md`

## Estado actual (última sesión: 2026-09-16)

- ✅ **Onboarding con documentos (nuevo, 2026-09-16)**: nueva pestaña "Documentos" en el menú del instructor (`DocumentsPage.tsx`) — tarjeta por documento (libreta, cédula verde, seguro) con fecha de vencimiento donde aplica y subida de foto, mismo patrón visual que `ExamsPage.tsx`. `ProfilePage.tsx` (Inicio) muestra un banner si `VerificationStatus !== "approved"` con link directo a Documentos, y ahora captura el error 403 si intenta ponerse online sin estar verificado. Admin tiene su propia pestaña "Documentos" (`DocumentsReviewPage.tsx`) para aprobar/rechazar.
- ✅ **Credenciales de admin para pruebas**: `admin_test_1789441604@test.com` / `adminpass123` (creado con `cmd/seedadmin` en el backend).
- Fix chico de esta sesión: el `<select>` de horas/minutos en `AvailabilityEditor` mostraba solo un dígito con la flechita nativa del navegador encima — le saqué el `appearance` nativo y ajusté el ancho.

## Estado actual (última sesión: 2026-09-15)

- ✅ **Agenda del instructor rediseñada + excepciones (nuevo, 2026-09-15)**: `AvailabilityEditor.tsx` pasó de una lista plana a una semana completa siempre visible (7 días, aunque estén vacíos: "Sin horarios") con los bloques como chips y "+ Agregar" contextual por día (selectores de hora en 24hs, no el `<input type="time">` nativo que renderizaba AM/PM y confundía). Nació de dos mockups que el usuario aprobó antes de pasarlos a código.
  - Sección nueva "Excepciones puntuales" debajo: día libre u horario especial para una fecha concreta, que pisa la agenda semanal solo esa fecha (resuelve el caso "el lunes que viene trabajo medio día" sin tocar el patrón de siempre). Lista de excepciones cargadas con "Quitar".
- ✅ Login/registro, rutas protegidas por rol (`student` / `instructor`).
- ✅ **Instructor**: completar perfil (vehículo, tarifa), toggle online/offline con geolocalización, heartbeat de ubicación cada 10s mientras está online (necesario para que el backend no lo dé de baja del matching — ver TTL de 30s en el backend), modal de oferta entrante con cuenta regresiva (`features/instructor/IncomingOfferModal.tsx`).
- ✅ **Estudiante**: mapa (Leaflet) con instructores cercanos, click para fijar el punto de encuentro, botón "Pedir clase ahora", banners de estado (buscando / asignado / sin instructores) alimentados por WebSocket en tiempo real.
- ✅ Conexión WebSocket persistente (`src/ws/socketManager.ts` + `src/ws/useWebSocket.ts`), montada una sola vez en `App.tsx`, con **reconexión automática** (2s) y reenvío de ubicación al reconectar — corrige un bug real donde el instructor se caía del matching en silencio si el socket se cortaba.
- ✅ **Reserva de horarios fijos, v2 (2026-09-14)**: rediseñado a partir de un mockup que el usuario aprobó (`features/student/SlotSearch.tsx`). Chips de día (Hoy/Mañana/+4), filtro de instructor, tarjetas agrupadas por instructor (avatar con iniciales, rating si tiene, vehículo, precio) con los horarios como chips. Selección **múltiple** — se pueden elegir horarios de distintos días antes de confirmar — con una barra flotante ("N horarios seleccionados") que abre un modal de revisión: ahí se puede sacar alguno antes de confirmar todo junto. Si algún horario se lo ganaron mientras tanto, avisa cuáles fallaron y confirma el resto.
  - Decisión de diseño del usuario: **esta vista no lleva mapa**. El mapa se reserva para el día de la clase (seguimiento en vivo instructor↔estudiante llegando al punto de encuentro) — eso todavía no existe como pantalla, es una nota para una fase futura, no se construyó.
  - `instructor/AvailabilityEditor.tsx` (agenda semanal del instructor) sigue igual. Se borró `InstructorSlotsModal.tsx` (el modal por-instructor de la v1), reemplazado por el flujo de carrito de `SlotSearch.tsx`. El mapa (`MapView.tsx`) volvió a ser solo para "pedir clase ahora" — se le sacó el botón "Ver horarios" que tenía en el popup.
- ✅ **Pago al confirmar reserva (nuevo, 2026-09-14)**: el modal de `SlotSearch.tsx` ahora muestra el desglose (subtotal instructores + fee de servicio 10% = total) antes de confirmar, y el botón dice "Confirmar y pagar $X". Al confirmar, cada reserva ya cobra automáticamente en el backend (simulado, aprobado al instante — ver README del backend); el toast final muestra el monto real cobrado tomado de la respuesta de la API, no del cálculo local. `FEE_PREVIEW_PERCENT` en `SlotSearch.tsx` es solo para pintar el desglose antes de confirmar — el backend es la fuente de verdad del monto real.
- ✅ **Menú del estudiante + exámenes (nuevo, 2026-09-15)**: `StudentLayout.tsx` agrega un sub-nav (Inicio / Reservas / Clases / Pagos / Exámenes) vía rutas anidadas de React Router bajo `/student`. Páginas nuevas: `ReservationsPage` (próximas, con cancelar), `LessonHistoryPage` (completadas + horas totales), `PaymentsPage` (historial con desglose instructor/fee), `ExamsPage` (progreso hacia cada examen, subir foto cuando está habilitado, ver estado de la revisión).
  - Nuevo rol con UI propia: **admin** — `features/admin/ExamsReviewPage.tsx` en `/admin`, cola de exámenes pendientes con aprobar/rechazar (rechazar pide motivo con un `prompt()`, simple a propósito). `App.tsx` ahora redirige admin correctamente tanto en `Home()` como al loguearse (antes solo distinguía instructor/estudiante — un admin logueándose rebotaba una vez antes de corregirse solo).
  - Alcance decidido con el usuario: este menú es **solo para el estudiante** por ahora; el instructor no tenía uno — hasta el punto siguiente.
- ✅ **Menú del instructor + wallet + medio de cobro (nuevo, 2026-09-15)**: `InstructorLayout.tsx` agrega sub-nav (Inicio / Reservas / Clases dadas / Wallet / Medio de cobro). `InstructorReservationsPage` ahora sí deja **empezar clase** y **completar** desde la UI (antes solo existían los endpoints, sin pantalla) además de cancelar; `InstructorLessonsGivenPage` es el historial con horas y $ generados; `WalletPage` muestra saldo por cobrar + historial de liquidaciones; `PayoutMethodPage` es el form de banco/tipo de cuenta/número/titular/CI.
  - Ambas páginas de reservas/clases del instructor resuelven el nombre del estudiante vía el nuevo `GET /api/users/:id` (`api/users.ts`) — antes no había forma de mostrarlo.
  - Admin ahora tiene su propio sub-nav (`AdminLayout.tsx`: Exámenes / Liquidaciones). `PayoutsPage` lista instructores con saldo pendiente y un botón "Marcar como pagado" — **no transfiere plata real**, es solo el registro de que el admin ya transfirió a mano.
- ⬜ Nada de esto se probó a mano en el navegador esta sesión (sin herramienta de browser disponible) — el matching en tiempo real y la reconexión de WS sí los confirmó el usuario en dos navegadores reales en una sesión anterior. El resto (reserva de horarios, pagos, menú, exámenes) se verificó con `npm run build` + pruebas de la API por curl, no clickeado en pantalla.

## Cómo levantar

```bash
npm install     # si es la primera vez o cambiaron dependencias
npm run dev     # sirve en :5173
```

Requiere que el backend esté corriendo en `:8080` (ver README del backend). La URL de la API se configura en `.env` (`VITE_API_URL`, ya seteado a `http://localhost:8080`).

## Verificar que todo compila

```bash
npm run build   # tsc -b && vite build
```

## Estructura

- `src/api/`: clientes REST (`auth`, `instructors`, `lessons`, `availability`, `payments`, `exams`, `payouts`, `users`) + `client.ts` (interceptor de JWT).
- `src/ws/`: conexión WebSocket persistente (con reconexión) y hook de integración con el store.
- `src/store/`: zustand — `authStore` (sesión), `lessonStore` (oferta pendiente / lección actual).
- `src/features/auth|student|instructor|admin/`: páginas por rol. `StudentLayout.tsx`/`InstructorLayout.tsx`/`AdminLayout.tsx` son los shells con sub-nav de cada rol; `instructor/AvailabilityEditor.tsx` (agenda) y `student/SlotSearch.tsx` (buscar + reservar, con carrito) son la reserva de horarios; `admin/ExamsReviewPage.tsx` y `admin/PayoutsPage.tsx` son las dos colas de trabajo del admin.

## Para retomar

1. Probar todo a mano en el navegador — nada de lo de esta sesión (menús, reservas múltiples, pagos, exámenes, wallet, liquidaciones) se clickeó de verdad todavía.
2. Fase futura (idea del usuario, no construida): una pantalla de "día de la clase" con mapa en vivo para que instructor y estudiante se sigan mutuamente llegando al punto de encuentro — el mapa de `StudentHome` hoy solo sirve para "pedir clase ahora".
3. Cuando el backend tenga pagos reales: armar la pantalla de checkout/redirect (hoy el pago simulado no necesita redirect, se resuelve inline).
4. `PayoutMethodPage` no valida el formato del número de cuenta ni la cédula — si hace falta, agregar validación antes de que un instructor cargue datos mal.
