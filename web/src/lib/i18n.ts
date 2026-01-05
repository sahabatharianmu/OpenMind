import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    debug: true,
    fallbackLng: 'en',
    interpolation: {
      escapeValue: false, // not needed for react as it escapes by default
    },
    resources: {
      en: {
        translation: {
           "common": {
                "currency": "Currency",
                "price": "Price",
                "save": "Save",
                "cancel": "Cancel",
                "plans": "Subscription Plans",
                "manage_pricing": "Manage your product pricing and packages.",
                "edit_plan": "Edit Plan",
                "no_plans": "No plans found. Create one to get started.",
                "active": "Active",
                "inactive": "Inactive",
                "create_plan": "Create Plan",
                "name": "Name",
                "description": "Description",
                "limits": "Limits",
                "patients": "Patients",
                "clinicians": "Clinicians",
                "unlimited": "Unlimited"
           }
        }
      },
      id: {
        translation: {
             "common": {
                "currency": "Mata Uang",
                "price": "Harga",
                "save": "Simpan",
                "cancel": "Batal",
                "plans": "Paket Berlangganan",
                "manage_pricing": "Kelola harga dan paket produk Anda.",
                "edit_plan": "Edit Paket",
                "no_plans": "Tidak ada paket ditemukan. Buat satu untuk memulai.",
                "active": "Aktif",
                "inactive": "Tidak Aktif",
                "create_plan": "Buat Paket",
                "name": "Nama",
                "description": "Deskripsi",
                "limits": "Batasan",
                "patients": "Pasien",
                "clinicians": "Klinisi",
                "unlimited": "Tak Terbatas"
           }
        }
      }
    }
  });

export default i18n;
