"use client";

import { useState } from "react";

type QRResponse = {
  payload: string;
  object: string;
  qr_url: string;
  expires_in: number;
  created_utc: string;
};

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8006";

export default function Home() {
  const [mode, setMode] = useState("promptpay");
  const [promptPayID, setPromptPayID] = useState("");
  const [amount, setAmount] = useState("");
  const [billerID, setBillerID] = useState("");
  const [reference1, setReference1] = useState("");
  const [reference2, setReference2] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<QRResponse | null>(null);

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);
    setResult(null);
    setLoading(true);

    try {
      const body: Record<string, string | undefined> = {
        mode,
        amount: amount.trim() === "" ? undefined : amount,
      };

      if (mode === "promptpay") {
        body.promptpay_id = promptPayID;
      }

      if (mode === "biller") {
        body.biller_id = billerID;
        body.reference1 = reference1;
        body.reference2 = reference2.trim() === "" ? undefined : reference2;
      }

      const response = await fetch(`${API_BASE_URL}/api/qr`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(body),
      });

      if (!response.ok) {
        const data = (await response.json().catch(() => null)) as {
          error?: string;
        } | null;
        throw new Error(data?.error ?? "Request failed");
      }

      const data = (await response.json()) as QRResponse;
      setResult(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Something went wrong");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center px-6 py-16">
      <main className="w-full max-w-5xl">
        <div className="grid gap-10 rounded-3xl border border-ring bg-card/80 p-8 shadow-[0_40px_120px_-60px_rgba(0,0,0,0.4)] backdrop-blur sm:p-12 lg:grid-cols-[1.1fr_0.9fr]">
          <section className="space-y-6">
            <div className="space-y-3">
              <p className="text-sm font-semibold uppercase tracking-[0.3em] text-accent">
                PromptPay QR Studio
              </p>
              <h1 className="text-4xl font-semibold leading-tight text-foreground sm:text-5xl">
                Generate PromptPay or bill payment QR codes.
              </h1>
              <p className="max-w-xl text-lg text-muted">
                Choose the QR type, fill the fields, and the Go backend will
                render the QR image and store it in MinIO for mobile banking.
              </p>
            </div>

            <form className="space-y-5" onSubmit={handleSubmit}>
              <div className="space-y-2">
                <label className="text-sm font-medium text-foreground">
                  QR Type
                </label>
                <select
                  className="w-full rounded-xl border border-ring bg-white/70 px-4 py-3 text-base text-foreground shadow-sm focus:border-accent focus:outline-none"
                  value={mode}
                  onChange={(event) => setMode(event.target.value)}
                >
                  <option value="promptpay">
                    PromptPay (phone / national ID)
                  </option>
                  <option value="biller">
                    Bill Payment (biller ID + reference)
                  </option>
                </select>
              </div>

              {mode === "promptpay" ? (
                <div className="space-y-2">
                  <label className="text-sm font-medium text-foreground">
                    PromptPay ID
                  </label>
                  <input
                    className="w-full rounded-xl border border-ring bg-white/70 px-4 py-3 text-base text-foreground shadow-sm focus:border-accent focus:outline-none"
                    placeholder="Phone (10 digits) or National ID (13 digits)"
                    value={promptPayID}
                    onChange={(event) => setPromptPayID(event.target.value)}
                    required
                    autoComplete="off"
                  />
                </div>
              ) : (
                <div className="grid gap-4 sm:grid-cols-2">
                  <div className="space-y-2 sm:col-span-2">
                    <label className="text-sm font-medium text-foreground">
                      Biller ID
                    </label>
                    <input
                      className="w-full rounded-xl border border-ring bg-white/70 px-4 py-3 text-base text-foreground shadow-sm focus:border-accent focus:outline-none"
                      placeholder="15-digit biller ID"
                      value={billerID}
                      onChange={(event) => setBillerID(event.target.value)}
                      required
                      autoComplete="off"
                    />
                  </div>
                  <div className="space-y-2">
                    <label className="text-sm font-medium text-foreground">
                      Reference 1
                    </label>
                    <input
                      className="w-full rounded-xl border border-ring bg-white/70 px-4 py-3 text-base text-foreground shadow-sm focus:border-accent focus:outline-none"
                      placeholder="Required"
                      value={reference1}
                      onChange={(event) => setReference1(event.target.value)}
                      required
                      autoComplete="off"
                    />
                  </div>
                  <div className="space-y-2">
                    <label className="text-sm font-medium text-foreground">
                      Reference 2 (optional)
                    </label>
                    <input
                      className="w-full rounded-xl border border-ring bg-white/70 px-4 py-3 text-base text-foreground shadow-sm focus:border-accent focus:outline-none"
                      placeholder="Optional"
                      value={reference2}
                      onChange={(event) => setReference2(event.target.value)}
                      autoComplete="off"
                    />
                  </div>
                </div>
              )}
              <div className="space-y-2">
                <label className="text-sm font-medium text-foreground">
                  Amount (optional)
                </label>
                <input
                  className="w-full rounded-xl border border-ring bg-white/70 px-4 py-3 text-base text-foreground shadow-sm focus:border-accent focus:outline-none"
                  placeholder="100.00"
                  value={amount}
                  onChange={(event) => setAmount(event.target.value)}
                  autoComplete="off"
                />
              </div>
              <button
                className="inline-flex w-full items-center justify-center gap-3 rounded-full bg-accent px-6 py-3 text-base font-semibold text-white transition hover:bg-[#0a5a41] disabled:cursor-not-allowed disabled:opacity-60"
                type="submit"
                disabled={loading}
              >
                {loading ? "Generating QR..." : "Generate QR"}
              </button>
            </form>

            {error ? (
              <div className="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
                {error}
              </div>
            ) : null}
          </section>

          <section className="flex flex-col items-center justify-center gap-6 rounded-2xl border border-dashed border-ring bg-white/60 p-6 text-center">
            <div className="space-y-2">
              <p className="text-sm font-semibold uppercase tracking-[0.2em] text-muted">
                Preview
              </p>
              <h2 className="text-2xl font-semibold text-foreground">
                Your QR code appears here
              </h2>
            </div>
            {result ? (
              <div className="space-y-4">
                <img
                  className="mx-auto h-56 w-56 rounded-2xl border border-ring bg-white p-3 shadow"
                  src={result.qr_url}
                  alt="PromptPay QR"
                />
                <div className="space-y-1 text-left text-xs text-muted">
                  <p>Object: {result.object}</p>
                  <p>Expires: {result.expires_in}s</p>
                  <p>Created: {result.created_utc}</p>
                  <p className="break-all">Payload: {result.payload}</p>
                </div>
              </div>
            ) : (
              <p className="text-sm text-muted">
                Submit the form to generate a QR code preview.
              </p>
            )}
          </section>
        </div>
      </main>
    </div>
  );
}
