import { useEffect, useMemo, useState } from 'react'
import { pathPrefix } from '../env'
import methods, {
  WorkerMethod,
  WorkerMethodName,
  WorkerParam,
  WorkerResult,
} from './methods'

type RecvMessage<N extends WorkerMethodName> = {
  data: WorkerResult<N>
}

type NextRun<N extends WorkerMethodName> = {
  arg: WorkerParam<N>
}

type ChangeCallback<N extends WorkerMethodName> = (
  result: WorkerResult<N>,
) => void

type Post<N extends WorkerMethodName> = {
  method: N
  arg: WorkerParam<N>
}

// StubWorker does work after a setTimeout, but in the main thread.
class StubWorker<N extends WorkerMethodName> {
  constructor(methodName: N) {
    this.method = methods[methodName]
  }

  private method: WorkerMethod<N>
  private _timeout: ReturnType<typeof setTimeout> | undefined
  onmessage: (e: RecvMessage<N>) => void = (): void => {}

  postMessage = (data: Post<N>): void => {
    this._timeout = setTimeout(() => {
      this.onmessage({
        // typescript calculates the incorrect type for the method argument
        /* eslint-disable @typescript-eslint/no-explicit-any */
        data: this.method(data.arg as any) as WorkerResult<N>,
      })
    })
  }

  terminate = (): void => {
    if (!this._timeout) return
    clearTimeout(this._timeout)
  }
}

class Runner<N extends WorkerMethodName> {
  constructor(methodName: N, onChange: ChangeCallback<N>) {
    this.methodName = methodName
    this.onChange = onChange
  }

  private methodName: N
  private worker: Worker | StubWorker<N> | null = null
  private next: NextRun<N> | null = null
  private onChange: ChangeCallback<N>
  private isBusy = false

  private _initWorker = (): Worker | StubWorker<N> => {
    const w = window.Worker
      ? new Worker(`${pathPrefix}/static/worker.js`)
      : new StubWorker(this.methodName)

    w.onmessage = (e: RecvMessage<N>) => {
      this.isBusy = false
      this.onChange(e.data)
      this._send()
    }

    return w
  }

  private _send = (): void => {
    if (!this.next) return
    if (!this.worker) {
      this.worker = this._initWorker()
    }
    if (this.isBusy) return
    this.worker.postMessage({ method: this.methodName, arg: this.next.arg })
    this.isBusy = true
    this.next = null
  }

  run = (arg: WorkerParam<N>): void => {
    this.next = { arg }
    this._send()
  }

  shutdown = (): void => {
    if (!this.worker) return
    this.worker.terminate()
    this.worker = null
  }
}

export function useWorker<N extends WorkerMethodName>(
  methodName: N,
  arg: WorkerParam<N>,
  def: WorkerResult<N>,
): WorkerResult<N> {
  if (!(methodName in methods)) {
    throw new Error(`method must be a valid method from app/worker/methods.ts`)
  }

  const [result, setResult] = useState(def)
  const [worker, setWorker] = useState<Runner<N> | null>(null)

  useEffect(() => {
    const w = new Runner<N>(methodName, setResult)
    setWorker(w)
    return w.shutdown
  }, [])

  useMemo(() => {
    if (!worker) return
    worker.run(arg)
  }, [worker, arg])

  return result
}
