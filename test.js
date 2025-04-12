// Kayıt testi için basit JavaScript kodu
// Bu dosyayı test.js olarak kaydedip Node.js ile çalıştırabilirsiniz.
// Node.js ile çalıştırmak için: node test.js

const fetch = require("node-fetch");

// Kullanıcı kaydı için POST isteği
async function testRegister() {
  const url = "http://localhost:8080/user/register";

  const userData = {
    email: "admin@hotmail.com",
    username: "Admin",
    password: "1234asd",
  };

  try {
    console.log("Kayıt isteği gönderiliyor...");
    const response = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(userData),
    });

    const data = await response.json();

    console.log("Durum Kodu:", response.status);
    console.log("Yanıt:", JSON.stringify(data, null, 2));

    if (response.ok) {
      console.log("Kayıt işlemi başarılı!");
    } else {
      console.log("Kayıt işlemi başarısız!");
    }
  } catch (error) {
    console.error("Hata:", error.message);
  }
}

// Fonksiyonu çağır
testRegister();
