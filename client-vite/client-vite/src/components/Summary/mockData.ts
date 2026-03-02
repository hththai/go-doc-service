import type { Category } from "@/api/categories";

export type Item = {
  itemName: string;
  itemQty: string;
  unitPrice: string;
  subTotal: string;
};

export type Purchase = {
  id: string;
  title: string;
  filename: string;
  fileUrl?: string;
  buyAt: string;
  buyFrom: string;
  buyPrice: string;
  items: Item[];
  categories?: Category[];
};

export const MONTH_NAMES = [
  "January",
  "February",
  "March",
  "April",
  "May",
  "June",
  "July",
  "August",
  "September",
  "October",
  "November",
  "December",
];

export function formatDate(dateStr: string) {
  if (!dateStr) return "—";
  // Append time component to avoid UTC-midnight timezone shift for date-only strings
  const normalized = dateStr.includes("T") ? dateStr : `${dateStr}T00:00:00`;
  const date = new Date(normalized);
  if (Number.isNaN(date.getTime())) return "—";
  return date.toLocaleDateString("en-AU", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  });
}

export function formatCurrency(amount: string | number | null | undefined) {
  const num = Number.parseFloat((amount ?? "").toString());
  if (Number.isNaN(num)) return "—";
  return `$${num.toLocaleString("en-AU", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })}`;
}

export const MOCK_PURCHASES: Purchase[] = [
  // 2024 – March
  {
    id: "2024-03-1",
    title: "Monthly Groceries",
    filename: "coles_mar2024.pdf",
    fileUrl: "https://www.africau.edu/images/general/sample.pdf",
    buyAt: "2024-03-05",
    buyFrom: "Coles",
    buyPrice: "112.30",
    items: [
      {
        itemName: "Pasta 500g",
        itemQty: "3",
        unitPrice: "2.50",
        subTotal: "7.50",
      },
      {
        itemName: "Tomato Sauce",
        itemQty: "2",
        unitPrice: "3.80",
        subTotal: "7.60",
      },
      {
        itemName: "Beef Mince 500g",
        itemQty: "4",
        unitPrice: "9.00",
        subTotal: "36.00",
      },
    ],
  },
  {
    id: "2024-03-2",
    title: "Gym Membership",
    filename: "gym_invoice_mar24.pdf",
    buyAt: "2024-03-01",
    buyFrom: "Fitness First",
    buyPrice: "79.99",
    items: [
      {
        itemName: "Monthly Membership",
        itemQty: "1",
        unitPrice: "79.99",
        subTotal: "79.99",
      },
    ],
  },

  // 2024 – June
  {
    id: "2024-06-1",
    title: "Mid-Year Stationery",
    filename: "officeworks_jun24.jpg",
    buyAt: "2024-06-12",
    buyFrom: "Officeworks",
    buyPrice: "68.50",
    items: [
      {
        itemName: "Notebook A5 (5pk)",
        itemQty: "1",
        unitPrice: "18.00",
        subTotal: "18.00",
      },
      {
        itemName: "Pens (10pk)",
        itemQty: "2",
        unitPrice: "8.00",
        subTotal: "16.00",
      },
      {
        itemName: "Highlighters",
        itemQty: "1",
        unitPrice: "6.50",
        subTotal: "6.50",
      },
    ],
  },
  {
    id: "2024-06-2",
    title: "Team Building Dinner",
    filename: "italian_bistro_jun24.jpg",
    fileUrl: "https://picsum.photos/seed/bistro/800/1060",
    buyAt: "2024-06-28",
    buyFrom: "Italian Bistro",
    buyPrice: "345.00",
    items: [
      {
        itemName: "Pasta Carbonara (x5)",
        itemQty: "5",
        unitPrice: "28.00",
        subTotal: "140.00",
      },
      {
        itemName: "Wood-fired Pizza (x3)",
        itemQty: "3",
        unitPrice: "32.00",
        subTotal: "96.00",
      },
      {
        itemName: "Wine Bottle",
        itemQty: "2",
        unitPrice: "45.00",
        subTotal: "90.00",
      },
    ],
  },

  // 2024 – September
  {
    id: "2024-09-1",
    title: "Home Office Upgrade",
    filename: "jbhifi_sep24.pdf",
    fileUrl: "https://www.africau.edu/images/general/sample.pdf",
    buyAt: "2024-09-10",
    buyFrom: "JB Hi-Fi",
    buyPrice: "499.00",
    items: [
      {
        itemName: '27" Monitor',
        itemQty: "1",
        unitPrice: "399.00",
        subTotal: "399.00",
      },
      {
        itemName: "Monitor Stand",
        itemQty: "1",
        unitPrice: "59.00",
        subTotal: "59.00",
      },
      {
        itemName: "HDMI Cable 2m",
        itemQty: "1",
        unitPrice: "29.00",
        subTotal: "29.00",
      },
    ],
  },
  {
    id: "2024-09-2",
    title: "Spring Groceries",
    filename: "woolworths_sep24.pdf",
    buyAt: "2024-09-22",
    buyFrom: "Woolworths",
    buyPrice: "93.75",
    items: [
      {
        itemName: "Salmon Fillet 400g",
        itemQty: "2",
        unitPrice: "14.00",
        subTotal: "28.00",
      },
      {
        itemName: "Salad Mix",
        itemQty: "3",
        unitPrice: "4.50",
        subTotal: "13.50",
      },
      {
        itemName: "Orange Juice 2L",
        itemQty: "2",
        unitPrice: "5.50",
        subTotal: "11.00",
      },
    ],
  },

  // 2024 – December
  {
    id: "2024-12-1",
    title: "Christmas Party",
    filename: "catering_dec24.pdf",
    buyAt: "2024-12-20",
    buyFrom: "Fresh Catering Co.",
    buyPrice: "620.00",
    items: [
      {
        itemName: "Catering Platter (x4)",
        itemQty: "4",
        unitPrice: "85.00",
        subTotal: "340.00",
      },
      {
        itemName: "Soft Drinks Crate",
        itemQty: "3",
        unitPrice: "40.00",
        subTotal: "120.00",
      },
      {
        itemName: "Dessert Assortment",
        itemQty: "2",
        unitPrice: "60.00",
        subTotal: "120.00",
      },
    ],
  },
  {
    id: "2024-12-2",
    title: "Year-End Supplies",
    filename: "bunnings_dec24.jpg",
    buyAt: "2024-12-05",
    buyFrom: "Bunnings",
    buyPrice: "157.40",
    items: [
      {
        itemName: "Storage Boxes (6pk)",
        itemQty: "2",
        unitPrice: "39.00",
        subTotal: "78.00",
      },
      {
        itemName: "Label Maker",
        itemQty: "1",
        unitPrice: "49.90",
        subTotal: "49.90",
      },
    ],
  },

  // 2025 – January
  {
    id: "2025-01-1",
    title: "Weekly Groceries",
    filename: "invoice_woolworths_jan.pdf",
    fileUrl: "https://www.africau.edu/images/general/sample.pdf",
    buyAt: "2025-01-15",
    buyFrom: "Woolworths",
    buyPrice: "87.45",
    items: [
      {
        itemName: "Milk 2L",
        itemQty: "2",
        unitPrice: "3.50",
        subTotal: "7.00",
      },
      { itemName: "Bread", itemQty: "1", unitPrice: "4.20", subTotal: "4.20" },
      {
        itemName: "Chicken Breast 500g",
        itemQty: "3",
        unitPrice: "8.50",
        subTotal: "25.50",
      },
      {
        itemName: "Eggs 12pk",
        itemQty: "2",
        unitPrice: "6.00",
        subTotal: "12.00",
      },
    ],
  },
  {
    id: "2025-01-2",
    title: "Office Supplies",
    filename: "officeworks_receipt.jpg",
    buyAt: "2025-01-20",
    buyFrom: "Officeworks",
    buyPrice: "134.90",
    items: [
      {
        itemName: "A4 Paper Ream",
        itemQty: "2",
        unitPrice: "12.00",
        subTotal: "24.00",
      },
      {
        itemName: "Stapler",
        itemQty: "1",
        unitPrice: "18.90",
        subTotal: "18.90",
      },
      {
        itemName: "USB-C Hub",
        itemQty: "1",
        unitPrice: "69.00",
        subTotal: "69.00",
      },
    ],
  },

  // 2025 – February
  {
    id: "2025-02-1",
    title: "Team Lunch",
    filename: "thai_garden_receipt.jpg",
    fileUrl: "https://picsum.photos/seed/thaireceipt/800/1060",
    buyAt: "2025-02-03",
    buyFrom: "Thai Garden Restaurant",
    buyPrice: "210.00",
    items: [
      {
        itemName: "Pad Thai (x4)",
        itemQty: "4",
        unitPrice: "22.00",
        subTotal: "88.00",
      },
      {
        itemName: "Green Curry (x3)",
        itemQty: "3",
        unitPrice: "24.00",
        subTotal: "72.00",
      },
      {
        itemName: "Drinks",
        itemQty: "7",
        unitPrice: "7.14",
        subTotal: "50.00",
      },
    ],
  },
  {
    id: "2025-02-2",
    title: "Laptop Charger",
    filename: "amazon_invoice_feb.pdf",
    buyAt: "2025-02-10",
    buyFrom: "Amazon AU",
    buyPrice: "59.99",
    items: [
      {
        itemName: "USB-C 65W Charger",
        itemQty: "1",
        unitPrice: "59.99",
        subTotal: "59.99",
      },
    ],
  },
];
