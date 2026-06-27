"use client";

import { useEffect, useRef } from "react";
import * as THREE from "three";

// Starmind renders a 3D satellite constellation: a dark Earth ringed by equally
// spaced orbital planes of red satellites that flash comms beams to neighbours
// in line of sight. Ported from the slashnode-web landing background, but sized
// to its parent element and auto-orbiting (no page scroll here). Used as the
// full-screen "starmind" background; theme-aware via the .dark class + --primary.

// Coherent scale: Earth radius ≈ 6371 km, satellites at 300 km altitude (LEO).
const EARTH_R = 6;
const SHELL_R = EARTH_R * (1 + 300 / 6371); // ≈ 6.28 — same altitude for all
const INC = 0.93; // ~53° inclination, identical for every orbit
const SPEED = 0.22; // identical speed & direction
const BEAM_LIFE = 0.28; // seconds — fast flash
const NEAR = 1.5; // inter-orbit link range

export function Starmind({
  className,
  mini = false,
}: {
  className?: string;
  mini?: boolean;
}) {
  const mountRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const mount = mountRef.current;
    if (!mount) return;

    const PLANES = mini ? 10 : 22;
    const SATS_PER_PLANE = mini ? 10 : 16;
    const N = PLANES * SATS_PER_PLANE;
    const BEAM_POOL = mini ? 10 : 26;

    const reduce = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

    const scene = new THREE.Scene();
    const camera = new THREE.PerspectiveCamera(50, 1, 0.1, 200);
    const camRadius = 22;

    const renderer = new THREE.WebGLRenderer({ alpha: true, antialias: true });
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    renderer.setClearColor(0x000000, 0);
    mount.appendChild(renderer.domElement);

    const group = new THREE.Group();
    scene.add(group);

    // ---- Earth: an opaque globe that occludes the far-side satellites ----
    const earthFill = new THREE.Mesh(
      new THREE.SphereGeometry(EARTH_R, 64, 48),
      new THREE.MeshBasicMaterial({ color: 0x0c1424 }),
    );
    const earthWire = new THREE.LineSegments(
      new THREE.WireframeGeometry(new THREE.SphereGeometry(EARTH_R * 1.003, 24, 18)),
      new THREE.LineBasicMaterial({ color: 0x9fb3d9, transparent: true, opacity: 0.28 }),
    );
    group.add(earthFill, earthWire);

    // ---- orbital planes: same inclination, different RAAN ----
    const planeM: THREE.Matrix4[] = [];
    const orbitMat = new THREE.LineBasicMaterial({ color: 0x9fb3d9, transparent: true, opacity: 0.09 });
    for (let p = 0; p < PLANES; p++) {
      const node = (p / PLANES) * Math.PI * 2;
      const m = new THREE.Matrix4().makeRotationY(node).multiply(new THREE.Matrix4().makeRotationX(INC));
      planeM.push(m);
      const curve = new THREE.EllipseCurve(0, 0, SHELL_R, SHELL_R, 0, Math.PI * 2, false, 0);
      const pts = curve.getPoints(120).map((q) => new THREE.Vector3(q.x, q.y, 0).applyMatrix4(m));
      group.add(new THREE.LineLoop(new THREE.BufferGeometry().setFromPoints(pts), orbitMat));
    }

    // ---- satellites: many tiny red dots ----
    const positions = new Float32Array(N * 3);
    const satPos: THREE.Vector3[] = [];
    const planeOf = new Int32Array(N);
    const phase = new Float32Array(N);
    for (let p = 0, k = 0; p < PLANES; p++) {
      for (let s = 0; s < SATS_PER_PLANE; s++, k++) {
        planeOf[k] = p;
        phase[k] = (s / SATS_PER_PLANE) * Math.PI * 2;
        satPos.push(new THREE.Vector3());
      }
    }
    const satGeo = new THREE.BufferGeometry();
    satGeo.setAttribute("position", new THREE.BufferAttribute(positions, 3));
    const satMat = new THREE.PointsMaterial({ color: 0xe5484d, size: 0.07, sizeAttenuation: true, depthTest: true });
    group.add(new THREE.Points(satGeo, satMat));

    const tmp = new THREE.Vector3();
    function placeSats(t: number) {
      for (let i = 0; i < N; i++) {
        const ang = phase[i] - SPEED * t;
        tmp.set(Math.cos(ang) * SHELL_R, Math.sin(ang) * SHELL_R, 0).applyMatrix4(planeM[planeOf[i]]);
        satPos[i].copy(tmp);
        positions[i * 3] = tmp.x;
        positions[i * 3 + 1] = tmp.y;
        positions[i * 3 + 2] = tmp.z;
      }
      satGeo.getAttribute("position").needsUpdate = true;
    }

    // segment A–B blocked by the planet? (line-of-sight test)
    const EARTH_R2 = EARTH_R * 0.998 * (EARTH_R * 0.998);
    function blocked(a: THREE.Vector3, b: THREE.Vector3): boolean {
      const dx = b.x - a.x, dy = b.y - a.y, dz = b.z - a.z;
      const len2 = dx * dx + dy * dy + dz * dz;
      if (len2 === 0) return false;
      let t = -(a.x * dx + a.y * dy + a.z * dz) / len2;
      t = Math.max(0, Math.min(1, t));
      const cx = a.x + dx * t, cy = a.y + dy * t, cz = a.z + dz * t;
      return cx * cx + cy * cy + cz * cz < EARTH_R2;
    }

    // ---- red comms beams between satellites in different orbits, in sight ----
    type Beam = { line: THREE.LineSegments; mat: THREE.LineBasicMaterial; life: number; a: number; b: number };
    const beams: Beam[] = [];
    for (let i = 0; i < BEAM_POOL; i++) {
      const g = new THREE.BufferGeometry();
      g.setAttribute("position", new THREE.BufferAttribute(new Float32Array(6), 3));
      const mat = new THREE.LineBasicMaterial({
        color: 0xe5484d,
        transparent: true,
        opacity: 0,
        blending: THREE.AdditiveBlending,
        depthTest: true,
        depthWrite: false,
      });
      const line = new THREE.LineSegments(g, mat);
      line.visible = false;
      group.add(line);
      beams.push({ line, mat, life: 0, a: 0, b: 0 });
    }

    // Theme-aware Earth + primary-coloured dots/beams; re-applied on theme toggle.
    const rootEl = document.documentElement;
    function applyTheme() {
      const isDark = rootEl.classList.contains("dark");
      (earthFill.material as THREE.MeshBasicMaterial).color.set(isDark ? 0x0c1424 : 0xe6ecf6);
      (earthWire.material as THREE.LineBasicMaterial).color.set(isDark ? 0x9fb3d9 : 0x6f8bc0);
      orbitMat.color.set(isDark ? 0x9fb3d9 : 0x6f8bc0);
      const primary = getComputedStyle(rootEl).getPropertyValue("--primary").trim() || "#e5484d";
      satMat.color.set(primary);
      for (const b of beams) b.mat.color.set(primary);
    }
    applyTheme();
    const themeObs = new MutationObserver(applyTheme);
    themeObs.observe(rootEl, { attributes: true, attributeFilter: ["class", "style"] });

    function spawnBeam(beam: Beam) {
      const a = Math.floor(Math.random() * N);
      const cand: number[] = [];
      for (let j = 0; j < N; j++) {
        if (planeOf[j] === planeOf[a]) continue; // must be a DIFFERENT orbit
        if (satPos[a].distanceTo(satPos[j]) >= NEAR) continue; // close enough
        if (blocked(satPos[a], satPos[j])) continue; // and in line of sight
        cand.push(j);
      }
      if (cand.length === 0) return;
      beam.a = a;
      beam.b = cand[Math.floor(Math.random() * cand.length)];
      beam.life = BEAM_LIFE;
      beam.line.visible = true;
    }

    let spawnTimer = 0;
    function updateBeams(dt: number) {
      spawnTimer -= dt;
      if (spawnTimer <= 0) {
        spawnTimer = 0.025 + Math.random() * 0.05;
        for (const b of beams) {
          if (b.life <= 0) { spawnBeam(b); break; }
        }
      }
      for (const b of beams) {
        if (b.life <= 0) {
          if (b.line.visible) { b.line.visible = false; b.mat.opacity = 0; }
          continue;
        }
        b.life -= dt;
        const attr = b.line.geometry.getAttribute("position") as THREE.BufferAttribute;
        const pa = satPos[b.a];
        const pb = satPos[b.b];
        attr.setXYZ(0, pa.x, pa.y, pa.z);
        attr.setXYZ(1, pb.x, pb.y, pb.z);
        attr.needsUpdate = true;
        const k = Math.max(b.life / BEAM_LIFE, 0);
        b.mat.opacity = k * k * (0.7 + Math.random() * 0.3);
      }
    }

    // Size to the mount element (not the window) so it works full-screen and as
    // a small tile.
    function resize() {
      const r = mount!.getBoundingClientRect();
      const w = Math.max(1, r.width);
      const h = Math.max(1, r.height);
      camera.aspect = w / h;
      camera.updateProjectionMatrix();
      renderer.setSize(w, h, false);
      renderer.domElement.style.width = "100%";
      renderer.domElement.style.height = "100%";
    }
    resize();
    const ro = new ResizeObserver(resize);
    ro.observe(mount);

    let raf = 0;
    let last = performance.now();
    let elapsed = 0;
    let camAngle = 0;
    function frame() {
      const now = performance.now();
      const dt = Math.min((now - last) / 1000, 0.05);
      last = now;
      elapsed += dt;
      placeSats(elapsed);
      updateBeams(dt);
      // Gentle continuous orbit; slow north/south tilt (no page scroll here).
      camAngle += dt * 0.08;
      const camElev = Math.sin(elapsed * 0.05) * 0.45;
      const ring = Math.cos(camElev) * camRadius;
      camera.position.set(Math.cos(camAngle) * ring, Math.sin(camElev) * camRadius, Math.sin(camAngle) * ring);
      camera.lookAt(0, 0, 0);
      renderer.render(scene, camera);
      raf = requestAnimationFrame(frame);
    }

    if (reduce) {
      placeSats(0);
      camera.position.set(camRadius, 4, 0);
      camera.lookAt(0, 0, 0);
      renderer.render(scene, camera);
    } else {
      frame();
    }

    return () => {
      cancelAnimationFrame(raf);
      ro.disconnect();
      themeObs.disconnect();
      renderer.dispose();
      scene.traverse((obj) => {
        const o = obj as THREE.Mesh;
        if (o.geometry) o.geometry.dispose();
        const mat = o.material;
        if (Array.isArray(mat)) mat.forEach((m) => m.dispose());
        else if (mat) (mat as THREE.Material).dispose();
      });
      if (renderer.domElement.parentNode === mount) mount.removeChild(renderer.domElement);
    };
  }, [mini]);

  return <div ref={mountRef} aria-hidden className={className} />;
}
