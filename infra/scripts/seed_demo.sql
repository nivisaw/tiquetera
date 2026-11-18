-- Datos de demostración SOLO para entorno local (make seed-demo).
-- Crea 40 clientes con tiqueteras, recargas, consumos de las últimas semanas
-- y rechazos de hoy. Idempotente: no hace nada si ya existen.
DO $$
DECLARE
    nombres text[] := ARRAY[
        'Ana Pérez', 'Carlos Gómez', 'Luisa Fernández', 'Jorge Ramírez', 'Valentina Torres',
        'Andrés Castillo', 'Camila Rojas', 'Felipe Morales', 'Daniela Vargas', 'Santiago Herrera',
        'María José Ruiz', 'Juan David Ortiz', 'Paula Jiménez', 'Sebastián Castro', 'Laura Medina',
        'Diego Romero', 'Natalia Suárez', 'Mateo Navarro', 'Isabella Mendoza', 'Nicolás Guerrero',
        'Sofía Cárdenas', 'Alejandro Ríos', 'Mariana Peña', 'Julián Salazar', 'Gabriela Restrepo',
        'Tomás Acosta', 'Juliana Molina', 'David Cortés', 'Manuela Arango', 'Esteban Quintero',
        'Carolina Duarte', 'Samuel Pineda', 'Antonia Valencia', 'Miguel Ángel Osorio', 'Sara Londoño',
        'Daniel Escobar', 'Lucía Montoya', 'Emilio Zapata', 'Verónica Giraldo', 'Ricardo Patiño'
    ];
    admin_id   uuid;
    cajeros    uuid[];
    hoy        date := (now() AT TIME ZONE 'America/Bogota')::date;
    i          int;
    n          int;
    cid        uuid;
    tid        uuid;
    arrastre   int;
    vendidos   int;
    comprados  int;
    creada     timestamptz;
    dias_desde int;
    consumidos int;
BEGIN
    IF EXISTS (SELECT 1 FROM clientes WHERE telefono = '+573100000001') THEN
        RAISE NOTICE 'Datos demo ya cargados';
        RETURN;
    END IF;

    SELECT id INTO admin_id FROM usuarios WHERE rol = 'admin' ORDER BY created_at LIMIT 1;
    INSERT INTO usuarios (username, password_hash, nombre, rol)
    SELECT 'demo.cajero2', password_hash, 'Caja 2', 'cajero' FROM usuarios WHERE username = 'prueba.cajero'
    ON CONFLICT DO NOTHING;
    SELECT array_agg(id ORDER BY username) INTO cajeros FROM usuarios WHERE rol = 'cajero' AND activo;

    FOR i IN 1..40 LOOP
        INSERT INTO clientes (nombre, telefono, qr_token_hash, created_by, activo, created_at)
        VALUES (nombres[i], '+5731000000' || lpad(i::text, 2, '0'),
                sha256(convert_to('demo-' || i || '-' || clock_timestamp(), 'UTF8')),
                admin_id, i % 13 <> 0, now() - make_interval(days => 100 + i))
        RETURNING id INTO cid;

        arrastre := 0;
        IF i % 3 = 0 THEN
            arrastre := i % 4;
            INSERT INTO tiqueteras (cliente_id, almuerzos_comprados, almuerzos_vendidos, saldo,
                                    fecha_vencimiento, activa, created_by, created_at)
            VALUES (cid, 20, 20, arrastre, hoy - 40, false, admin_id, now() - make_interval(days => 70 + i));
        END IF;

        vendidos   := (ARRAY[10, 20, 20, 30])[1 + i % 4];
        comprados  := vendidos + arrastre;
        dias_desde := 1 + (i * 7) % 75;
        creada     := ((hoy - dias_desde) + time '11:30') AT TIME ZONE 'America/Bogota';
        consumidos := least(comprados - (i * 5) % 4, dias_desde);
        IF i % 5 = 0 THEN
            consumidos := least(comprados - 1, dias_desde);
        END IF;

        INSERT INTO tiqueteras (cliente_id, almuerzos_comprados, almuerzos_vendidos, saldo,
                                fecha_vencimiento, created_by, created_at)
        VALUES (cid, comprados, vendidos, comprados - consumidos,
                hoy + ((i * 11) % 50 - 8), admin_id, creada)
        RETURNING id INTO tid;

        FOR n IN 1..consumidos LOOP
            INSERT INTO consumos (tiquetera_id, cliente_id, cajero_id, saldo_resultante,
                                  idempotency_key, fecha_hora)
            VALUES (tid, cid, cajeros[1 + (i + n) % array_length(cajeros, 1)], comprados - n,
                    gen_random_uuid(),
                    creada + make_interval(days => n - 1)
                        + make_interval(hours => 1 + (i + n) % 3, mins => (i * n) % 60));
        END LOOP;
    END LOOP;

    INSERT INTO consumos_rechazados (cajero_id, motivo, fecha_hora)
    SELECT cajeros[1], m, now() - make_interval(mins => (k * 7)::int)
    FROM unnest(ARRAY['TIQUETERA_VENCIDA', 'SIN_SALDO', 'QR_INVALIDO']) WITH ORDINALITY AS r(m, k);
END
$$;
