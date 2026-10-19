# Clase Drive — Backend

Marketplace tipo Uber que conecta estudiantes con instructores de manejo en Uruguay. Go + Gin, PostgreSQL/PostGIS, Redis, WebSockets.

Plan completo del proyecto: `/Users/jesus.madrazo/.claude/plans/cheerful-enchanting-codd.md`

## Estado actual (última sesión: 2026-09-16)

- ✅ **Onboarding con verificación de documentos (nuevo, 2026-09-16)**: un instructor no puede ponerse online ni aparecer en `GET /api/instructors` (reserva de horarios) hasta que un admin apruebe sus 3 documentos: libreta de conducir, cédula verde/propiedad del vehículo, y seguro — cada uno con foto + fecha de vencimiento (opcional en cédula verde, obligatoria en libreta y seguro a nivel de negocio, aunque el backend no la fuerza).
  - `instructor_documents` (mismo patrón que `exams`: pending_review/approved/rejected, resubmit solo tras rechazo) + `instructor_profiles.verification_status` cacheado (pending/approved), recalculado en `DocumentService.Review` cada vez que un admin revisa un documento — evita tener que hacer un JOIN correlacionado cada vez que alguien intenta ir online o se lista para reserva.
  - Gate en dos puntos: `InstructorService.SetStatus` rechaza pasar a `online` si `verification_status != approved` (403 `ErrNotVerified`); `InstructorRepo.ListAll` solo devuelve instructores verificados.
  - `POST /api/instructors/me/documents` (multipart: `doc_type`, `expiry_date` opcional, `photo`), `GET /api/instructors/me/documents` (estado propio), `GET/POST /api/admin/documents/pending|:id/review`.
  - **Limitación conocida**: el vencimiento se chequea en el momento de calcular `verified` (`GetStatus`), pero `verification_status` cacheado en `instructor_profiles` solo se recalcula cuando un admin revisa un documento — si el seguro de un instructor ya verificado vence con el tiempo, `verification_status` sigue en `approved` hasta la próxima revisión manual. Demotion automática por vencimiento necesitaría un cron job; no se construyó.
  - Probado end-to-end con curl: instructor sin documentos → intento de ir online falla 403 → no aparece en `ListAll` → sube los 3 documentos → admin los aprueba → `verified: true` → va online → aparece en `ListAll`.

- ✅ **Excepciones puntuales a la agenda semanal (nuevo, 2026-09-15)**: la agenda recurrente (`instructor_availability`) es la base, pero un instructor puede pisarla para una fecha específica — día libre, u horario distinto solo ese día (ej. "el lunes que viene trabajo medio día") — vía `instructor_schedule_overrides` (una fila por instructor+fecha, `UNIQUE(instructor_id, date)`). Mientras exista una excepción para una fecha, reemplaza completamente el patrón recurrente para esa fecha, no se suma.
  - `POST/GET/DELETE /api/instructors/me/schedule-overrides[/:date]`.
  - `AvailabilityService.SlotsInRange` y `fitsSchedule` (usado tanto para calcular horarios como para validar un `BookSlot`) ahora chequean la excepción de la fecha primero, y solo caen al patrón semanal si no hay ninguna — así un intento de reservar un día marcado "libre" falla igual que si nunca hubiera existido el bloque recurrente.
  - Probado end-to-end con curl: instructor con disponibilidad 00:00–23:00 todos los días → seteo un horario especial 09:00–13:00 para una fecha puntual → los horarios calculados ese día bajan de 23 a 4 → seteo día libre para otra fecha → los horarios quedan en 0 y un intento de reserva directa a esa fecha devuelve 409 → listado y borrado de excepciones confirmados.

- ✅ **Fase 1 — Fundaciones**: auth JWT (register/login, roles student/instructor/admin), CRUD de perfil de instructor.
- ✅ **Fase 2 — Matching en tiempo real**: WebSocket Hub, Redis GEO para ubicación en vivo, `Matcher` que ofrece la clase al instructor más cercano (timeout 15s, reintento al siguiente), endpoints de lecciones (`create/get/list/start/complete/cancel/respond`), `GET /api/instructors/nearby`.
  - Probado end-to-end con un script temporal (ya borrado): instructor se conecta y acepta por WS, estudiante pide clase y recibe la asignación en tiempo real. Estado queda consistente en Postgres.
  - Bug real encontrado y corregido: las entradas de Redis GEO no expiran solas, así que un instructor que se cae sin togglear "offline" quedaba fantasma en el matching para siempre. Ahora `SearchNearby` filtra contra una clave con TTL de 30s (`internal/repository/redis/geo_repo.go`).
  - Bug real #2 encontrado y corregido (en el frontend): el WebSocket no se reconectaba solo si se cortaba (sleep de la laptop, blip de red, reinicio del backend), así que el heartbeat de ubicación del instructor moría en silencio y se caía del matching sin ningún aviso en pantalla. Ver `clase-drive-frontend/src/ws/socketManager.ts` (reconecta a los 2s, reenvía ubicación al reconectar).
- ✅ **Reserva de horarios fijos (nuevo, 2026-09-14)**: además del matching en tiempo real, el instructor puede publicar una agenda semanal recurrente (`instructor_availability`: día de la semana + hora de inicio/fin) y el estudiante puede reservar directamente un horario libre con un instructor específico — sin la carrera de aceptar/rechazar, queda confirmado al instante. Pensado para que el estudiante concentre sus horas con el mismo instructor en vez de que cada clase sea con alguien random.
  - `GET /api/instructors/:id/slots?days=N` calcula los horarios concretos disponibles expandiendo los bloques recurrentes contra lo ya reservado.
  - `POST /api/instructors/:id/book` reserva un horario. La concurrencia (dos estudiantes reservando el mismo slot a la vez) la resuelve un índice único parcial en Postgres (`idx_lesson_requests_instructor_slot`), no lógica de aplicación — probado con un script temporal que confirma el segundo intento falla con 409.
  - CRUD de disponibilidad: `POST/GET/DELETE /api/instructors/me/availability`.
  - Este flujo convive con el "pedir clase ahora" (`POST /api/lessons`); `lesson_requests.scheduled_at` es `NULL` para el modo en tiempo real y tiene fecha/hora para el modo agendado.
  - `GET /api/instructors/slots?date=YYYY-MM-DD&instructor_id=opcional` agrega los horarios de **todos** los instructores para una fecha (o de uno solo, si se filtra) — es la vista "elegí la fecha, mirá quién tiene lugar" que pidió el usuario, en vez de tener que entrar instructor por instructor. Ojo: recorre todos los instructores en el service layer (`SlotsForDate` en `availability_service.go`), aceptable para el volumen de un MVP pero no escala indefinidamente.
  - Bug de diseño encontrado y corregido: al principio la única forma de ver horarios era desde el mapa de instructores *online ahora*, lo cual no tiene sentido para reservar un horario de la semana que viene. Se separó en `GET /api/instructors` (lista completa, cualquier estado) vs `GET /api/instructors/nearby` (solo online, para el matching en tiempo real).
  - Bug encontrado al probar esto (no del feature en sí): mis reinicios del backend durante la sesión venían fallando en silencio (`pkill -f "cmd/api"` no siempre mata el binario que compila `go run`, y el puerto quedaba ocupado por el proceso viejo mientras el nuevo moría con "address already in use"). Si algo no refleja cambios recientes, verificar con `lsof -nP -iTCP:8080 -sTCP:LISTEN` que el PID sea el esperado antes de asumir que el código no anda.
- ✅ **Pagos simulados + fee de plataforma (nuevo, 2026-09-14)**: reservar un horario (`POST /api/instructors/:id/book`) ahora cobra automáticamente. Modelo: el instructor recibe exactamente lo que cotizó (`instructor_price_uyu`), la plataforma suma un fee del `PLATFORM_FEE_PERCENT` (10% por defecto, env var) **arriba** de ese precio — decisión explícita del usuario para que el instructor nunca vea menos de lo que puso. `payments.amount_uyu = instructor_price_uyu + platform_fee_uyu`.
  - Sin credenciales reales de Mercado Pago todavía, así que el "pago" lo resuelve `internal/provider/manual` (implementa `domain.PaymentProvider`, aprueba el checkout al instante). Cuando haya credenciales reales: escribir `internal/provider/mercadopago/` implementando la misma interfaz, apuntar `main.go` ahí, y pasar `autoApprove=false` en `service.NewPaymentService` (la aprobación pasaría a depender del webhook real). Nada en `AvailabilityService`, los handlers, o el frontend necesita cambiar.
  - Migración `0003_payment_split`: renombró `payments.mp_preference_id`/`mp_payment_id` a `provider_ref`/`provider_payment_id` (para que el nombre no ate la tabla a Mercado Pago) y agregó `instructor_price_uyu`/`platform_fee_uyu`.
  - `POST /api/instructors/:id/book` ahora devuelve `{lesson, payment}` en vez de solo la lección.
  - Nota: el pago se cobra al **reservar**, no al completar la clase — distinto del plan original (que cobraba al completar). Fue decisión explícita del usuario para el flujo de horarios fijos; el "pedir clase ahora" en tiempo real todavía no tiene cobro conectado.
- ✅ **Exámenes teórico/práctico (nuevo, 2026-09-15)**: reglas del usuario, implementadas literal. El examen teórico se habilita al completar `EXAM_THEORY_UNLOCK_LESSONS` clases (8 por defecto); el práctico al acumular `EXAM_PRACTICE_UNLOCK_MINUTES` minutos de clase completados (900 = 15h, ya coincide con la tarjeta de requisitos). Una vez habilitado, el estudiante sube una foto del resultado (`POST /api/exams/submit`, multipart) — queda `pending_review` hasta que un **admin** (no el instructor — así se decidió explícitamente, para evitar conflicto de interés ya que el instructor cobra por las clases) la aprueba o rechaza desde `GET /api/admin/exams/pending` + `POST /api/admin/exams/:id/review`. Rechazado permite volver a subir; aprobado o pendiente no (no se puede reintentar hasta que alguien lo rechace explícitamente).
  - Fotos guardadas en disco local (`UPLOADS_DIR`, default `./uploads`) y servidas vía `r.Static("/uploads", ...)` — **no es un volumen persistente**, se pierden en un redeploy sin volumen montado. Suficiente para MVP/dev, revisar antes de producción.
  - `LessonRepository.CompletionStats(studentID)` cuenta lecciones completadas + suma minutos — es la fuente de verdad de elegibilidad, no se guarda estado "unlocked" en la tabla `exams` (esa tabla solo tiene una fila una vez que el estudiante efectivamente sube algo).
  - Probado end-to-end con curl: sembré 8 lecciones completadas por SQL directo (no hay forma rápida de generarlas vía API en una prueba), confirmé que ambos exámenes se desbloquean, subí una foto real, la aprobé como admin, y confirmé que el estado del estudiante refleja `approved`.
- ✅ **Endpoint de pagos del estudiante**: `GET /api/payments/me` (antes solo existía `GetByLessonID`, faltaba listar todo el historial).
- ✅ **Wallet + medio de cobro + liquidación semanal del instructor (nuevo, 2026-09-15)**: `payments` ahora guarda `instructor_id` y `payout_id` (nullable) — un pago está "pendiente de cobro" mientras `payout_id IS NULL`. `GET /api/instructors/me/wallet` suma eso más el historial de liquidaciones ya pagadas. `PUT /api/instructors/me/payout-method` guarda banco/tipo de cuenta/número/titular/CI (tabla `instructor_payout_methods`).
  - **No hay transferencia bancaria real.** Es liquidación manual: el admin ve `GET /api/admin/payouts/pending` (saldo agregado por instructor), transfiere la plata por fuera de la app usando el medio de cobro guardado, y recién ahí llama `POST /api/admin/payouts/:instructorId`, que crea el `Payout` y marca esos pagos como liquidados (`MarkPaidForInstructor`). El período del payout se calcula del min/max `created_at` de los pagos pendientes incluidos.
  - Agregué `GET /api/users/:id` (cualquier usuario autenticado) para resolver el nombre de la contraparte en una reserva/lección — antes el instructor no tenía forma de ver el nombre del estudiante en su lista de reservas. Devuelve solo `{id, full_name}`, no es un directorio general de usuarios.
  - Probado end-to-end con curl: reservé una clase (genera pago con `instructor_id`), confirmé `pending_uyu` en la wallet, cargué el medio de cobro, liquidé como admin, y confirmé que la wallet pasa a `total_paid_uyu` con el payout en el historial.
- ⬜ **Fase 4 — Ratings y pulido**: no empezada.
- 🔧 **Pendiente (no bloqueante)**: al completar/cancelar una lección, devolver al instructor a `online` y reinsertarlo en Redis GEO (ahora mismo el instructor queda en `busy` para siempre tras la primera clase — hay que revisar `internal/transport/http/lesson_handler.go` `Complete`/`Cancel`).
- 📝 **Nota del usuario (2026-09-14)**: dijo que probablemente va a pedir cambiar la arquitectura del backend en algún momento — sin detalles todavía. No asumir nada, preguntar qué específicamente cuando lo mencione.

## Cómo levantar el entorno

Los puertos por defecto de Postgres/Redis (5432/6379) están ocupados por otro proyecto local (`liquidity-provider`), así que este proyecto usa **5433** y **6380**.

```bash
docker compose up -d              # Postgres+PostGIS en :5433, Redis en :6380
export $(grep -v '^#' .env | xargs)
migrate -path migrations -database "$DATABASE_URL" up   # solo la primera vez o tras cambios de esquema
go run ./cmd/api                  # sirve en :8080
```

`.env` ya existe localmente con un `JWT_SECRET` generado (no versionado). Si hace falta recrearlo, copiar `.env.example` y generar un secreto con `openssl rand -hex 32`.

Para crear un usuario admin:
```bash
go run ./cmd/seedadmin -email=admin@test.com -password=algoseguro -name="Admin"
```

## Verificar que todo compila

```bash
go build ./...
go vet ./...
```

## Estructura

Ver el plan para el detalle completo. Resumen rápido:
- `internal/domain`: entidades + interfaces de repos + `PaymentProvider` (interfaz de gateway de pago).
- `internal/provider/manual`: implementación simulada de `PaymentProvider` (aprueba al instante) — reemplazar por una real cuando haya credenciales.
- `internal/repository/postgres`: implementación con pgx.
- `internal/repository/redis`: `GeoRepo` (GEOADD/GEOSEARCH con filtro de TTL).
- `internal/service`: lógica de negocio — `auth_service.go`, `instructor_service.go`, `matching_service.go` (matching en tiempo real), `availability_service.go` (agenda semanal + reserva de horarios).
- `internal/transport/http`: handlers REST + middleware JWT/roles.
- `internal/transport/ws`: Hub de WebSocket + protocolo de eventos.

## Para retomar

1. Cuando haya credenciales de Mercado Pago (u otro proveedor): escribir `internal/provider/<nombre>/` implementando `domain.PaymentProvider`, apuntar `main.go` ahí en vez de `manualprovider.New()`, y pasar `autoApprove=false` — la aprobación pasaría a confirmarse por webhook en vez de al instante.
2. Cerrar el loop de "instructor vuelve a online tras completar/cancelar la clase" (ver pendiente arriba).
3. Fase 4: reviews + rating promedio, panel admin del lado instructor (hoy el menú nuevo es solo para el estudiante — el usuario no confirmó si también quiere uno para instructores).
4. Evaluar si "pedir clase ahora" (tiempo real) también debería cobrar, y en qué momento (al asignar vs. al completar).
5. Antes de producción: mover las fotos de exámenes a almacenamiento persistente (S3/GCS/volumen montado) — hoy es disco local efímero.
