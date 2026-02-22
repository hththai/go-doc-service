interface PurchaseInfoProps {
  values: {
    buyAt: string;
    buyFrom: string;
    buyPrice: string;
  };
  onChange: (name: string, value: string) => void;
}

const inputClass =
  "block min-w-0 grow bg-white py-1.5 pr-3 pl-1 text-base text-gray-900 placeholder:text-gray-400 focus:outline-none sm:text-sm/6";
const wrapperClass =
  "flex items-center rounded-md bg-white pl-3 outline-1 -outline-offset-1 outline-gray-300 focus-within:outline-2 focus-within:-outline-offset-2 focus-within:outline-indigo-600";
const labelClass = "block text-sm/6 font-medium text-gray-900";

export default function PurchaseInfo({
  values,
  onChange,
}: Readonly<PurchaseInfoProps>) {
  return (
    <div className="mt-6">
      <div className="bg-sky-500 text-white px-3 py-1 rounded-t-md">
        Purchase Info
      </div>
      <div className="border border-gray-200 rounded-b-md p-4 space-y-4">
        <div className="col-span-full">
          <label htmlFor="buyPrice" className={labelClass}>
            Purchase Price
          </label>
          <div className="mt-1">
            <div className={wrapperClass}>
              <input
                id="buyPrice"
                name="buyPrice"
                type="number"
                min="0"
                step="0.01"
                placeholder="0.00"
                value={values.buyPrice}
                onChange={(e) => onChange("buyPrice", e.target.value)}
                className={inputClass}
              />
            </div>
          </div>
        </div>

        <div className="col-span-full">
          <label htmlFor="buyFrom" className={labelClass}>
            Purchase From
          </label>
          <div className="mt-1">
            <div className={wrapperClass}>
              <input
                id="buyFrom"
                name="buyFrom"
                type="text"
                placeholder="e.g. Woolworths"
                value={values.buyFrom}
                onChange={(e) => onChange("buyFrom", e.target.value)}
                className={inputClass}
              />
            </div>
          </div>
        </div>

        <div className="col-span-full">
          <label htmlFor="buyAt" className={labelClass}>
            Purchase Date
          </label>
          <div className="mt-1">
            <div className={wrapperClass}>
              <input
                id="buyAt"
                name="buyAt"
                type="date"
                value={values.buyAt}
                onChange={(e) => onChange("buyAt", e.target.value)}
                className={inputClass}
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
