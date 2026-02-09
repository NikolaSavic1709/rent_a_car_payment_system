package server

import (
	"fmt"
	"net/http"
	"time"
	"bytes"
	"io"
	"strings"
	"strconv"
	"mime/multipart"
	"encoding/json"
	"psp_microservice/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Server) CardPaymentHandler(c *gin.Context) {
	tokenId := uuid.New()
	paymentURL := fmt.Sprintf("http://localhost:3001/card?tokenId=%s", tokenId)

	response := database.PaymentStartResponse{
		PaymentURL: paymentURL,
		TokenId:    tokenId,
		Token:      "token",
		TokenExp:   time.Now().Add(15 * time.Minute),
	}
	c.JSON(http.StatusOK, response)
}

func (s *Server) QrCodePaymentHandler(c *gin.Context, qrRef uint64) {
	tokenId := uuid.New()
	paymentURL := fmt.Sprintf("http://localhost:3001/qr?tokenId=%s", tokenId)
	fmt.Println("QR REF: ", qrRef)
	response := database.PaymentStartResponse{
		PaymentURL: paymentURL,
		TokenId:    tokenId,
		Token:      "token",
		TokenExp:   time.Now().Add(15 * time.Minute),
		QRRef:      qrRef,
	}
	c.JSON(http.StatusOK, response)
}

func (s *Server) CardDetailsHandler(c *gin.Context) {
	var req database.CardDetailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}
	fmt.Println(req.MerchantOrderId)
	paymentRequest, err := s.db.GetTransactionByMerchantOrderId(req.MerchantOrderId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error"})

	}
	fmt.Println("payment request")
	fmt.Println(paymentRequest.Amount)
	paymentRequest.CardNumber = req.CardNumber
	paymentRequest.ExpDate = req.ExpDate
	s.ForwardPaymentToBankGateway(paymentRequest)
	c.JSON(http.StatusOK, gin.H{"message": "Payment request forwarded"})
}

func (s *Server) QRCodeScanningHandler(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
        return
    }
    defer file.Close()

    // 2. Extract CardNumber and ExpDate from the same form-data body
    cardNumber := c.PostForm("CardNumber")
    expDateStr := c.PostForm("ExpDate")
    if cardNumber == "" || expDateStr == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "CardNumber and ExpDate are required"})
        return
    }
	parts := strings.Split(expDateStr, "/")
	if len(parts) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ExpDate format. Use MM/YY"})
		return
	}

	month, err := strconv.Atoi(parts[0])
	year, err2 := strconv.Atoi(parts[1])
	if err != nil || err2 != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ExpDate format. Use MM/YY"})
		return
	}

	expiryTimestamp := time.Date(
		2000+year,
		time.Month(month),
		1,
		0, 0, 0, 0,
		time.UTC,
	)


    // 2. Forward the file to NBS IPS upload API
    // NBS URL: https://nbs.rs/QRcode/api/qr/v1/upload [cite: 239]
    nbsResponse, err := s.ForwardToNBSUpload(file, header.Filename)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process QR code with NBS"})
        return
    }
	qrRef, err := strconv.ParseUint(nbsResponse.N.RO, 10, 64)
	if err != nil {
		// Handle error if the string contains non-numeric characters
		c.JSON(http.StatusBadRequest, gin.H{"error": "RO is not a valid number"})
		return
	}
	fmt.Println(qrRef)
	paymentRequest, err := s.db.GetTransactionByQRRef(qrRef)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error"})
		return 
	}
	fmt.Println("payment request")
	fmt.Println(paymentRequest)
	paymentRequest.CardNumber = cardNumber
	paymentRequest.ExpDate = expiryTimestamp
	s.ForwardPaymentToBankGateway(paymentRequest)
	c.JSON(http.StatusOK, gin.H{"message": "Payment request forwarded"})
}

func (s *Server) ForwardToNBSUpload(fileHeaderReader io.Reader, filename string) (*database.NBSUploadResponse, error) {
    // NBS API Endpoint for file upload 
    apiUrl := "https://nbs.rs/QRcode/api/qr/v1/upload"

    // Prepare a buffer to store the multipart request body
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    // Create the 'file' form field 
    part, err := writer.CreateFormFile("file", filename)
    if err != nil {
        return nil, err
    }

    // Copy the uploaded file content into the multipart form 
    _, err = io.Copy(part, fileHeaderReader)
    if err != nil {
        return nil, err
    }
    writer.Close()

    // Create the POST request [cite: 183, 239]
    req, err := http.NewRequest("POST", apiUrl, body)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", writer.FormDataContentType())

    // Execute the request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // Parse the NBS response [cite: 247]
    var nbsResp database.NBSUploadResponse
    if err := json.NewDecoder(resp.Body).Decode(&nbsResp); err != nil {
        return nil, err
    }

    // Check if NBS returned a success code (0 = OK) [cite: 247]
    if nbsResp.S.Code != 0 {
        return nil, fmt.Errorf("NBS Error: %s (Code: %d)", nbsResp.S.Desc, nbsResp.S.Code)
    }

    return &nbsResp, nil
}