/***********************************************
 * CONFIDENTIAL AND PROPRIETARY 
 * 
 * The source code and other information contained herein is the confidential and the exclusive property of
 * ZIH Corp. and is subject to the terms and conditions in your end user license agreement.
 * This source code, and any other information contained herein, shall not be copied, reproduced, published, 
 * displayed or distributed, in whole or in part, in any medium, by any means, for any purpose except as
 * expressly permitted under such license agreement.
 * 
 * Copyright ZIH Corp. 2012
 * 
 * ALL RIGHTS RESERVED
 ***********************************************/

package com.zebra.desktop.devdemo;

import java.awt.BorderLayout;
import java.awt.CardLayout;
import java.awt.Component;
import java.awt.Dimension;
import java.awt.event.FocusEvent;
import java.awt.event.FocusListener;
import java.awt.event.ItemEvent;
import java.awt.event.ItemListener;

import javax.swing.BorderFactory;
import javax.swing.BoxLayout;
import javax.swing.ButtonGroup;
import javax.swing.JButton;
import javax.swing.JComboBox;
import javax.swing.JFileChooser;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.JRadioButton;
import javax.swing.JTextField;
import javax.swing.filechooser.FileNameExtensionFilter;

import com.zebra.sdk.comm.Connection;
import com.zebra.sdk.comm.ConnectionException;
import com.zebra.sdk.comm.DriverPrinterConnection;
import com.zebra.sdk.comm.TcpConnection;
import com.zebra.sdk.comm.TlsConfig;
import com.zebra.sdk.comm.TlsConnection;
import com.zebra.sdk.comm.UsbConnection;
import com.zebra.sdk.printer.discovery.DiscoveredPrinterDriver;
import com.zebra.sdk.printer.discovery.UsbDiscoverer;

public class ConnectionCardPanel extends JPanel implements FocusListener {

    private static final long serialVersionUID = -1438813812046734210L;
    public JComboBox comboBox;
    public JComboBox usbPrinterList;
    public JTextField usbDirectAddressTextField;

    // Separate fields for Network and TLS
    public JTextField networkIpAddressTextField;
    public NumericTextField networkPortNumTextField;
    public JTextField tlsIpAddressTextField;
    public NumericTextField tlsPortNumTextField;
    private JTextField tlsCertPathTextField;
    private JButton tlsCertBrowseButton;

    private JRadioButton trustAllRadio;
    private JRadioButton trustJavaKsRadio;
    private JRadioButton trustCertFileRadio;

    private static final String tlsComboBoxLabel = "Network-TLS";
    private static final String networkComboBoxLabel = "Network";
    private static final String usbComboBoxLabel = "USB";
    private static final String usbDirectBoxLabel = "USB Direct";
    private String lastSelected = tlsComboBoxLabel;

    public ConnectionCardPanel() {
        super();
        addFocusListener(this);
        setFocusable(true);

        final JPanel cardPanel = new JPanel();
        JPanel comboBoxPanel = new JPanel();

        comboBox = new JComboBox(new String[] { tlsComboBoxLabel, networkComboBoxLabel, usbComboBoxLabel, usbDirectBoxLabel });
        comboBox.addItemListener(new ItemListener() {

            public void itemStateChanged(ItemEvent evt) {
                if (evt.getStateChange() == ItemEvent.SELECTED) {
                    String selected = (String) evt.getItem();
                    if (!selected.equals(lastSelected)) {
                        CardLayout cl = (CardLayout) (cardPanel.getLayout());
                        cl.show(cardPanel, selected);
                        lastSelected = selected;
                    }
                }
            }
        });
        comboBoxPanel.add(comboBox);

        cardPanel.setLayout(new CardLayout());

        // Create separate panels for each connection type
        JPanel tlsPanel = createTlsCard(); // Separate panel for TLS
        JPanel networkPanel = createNetworkCard();
        JPanel usbConnectivityPanel = createUsbCard();
        JPanel usbDirectConnectivityPanel = createUsbDirectCard();

        cardPanel.add(tlsPanel, tlsComboBoxLabel);
        cardPanel.add(networkPanel, networkComboBoxLabel);
        cardPanel.add(usbConnectivityPanel, usbComboBoxLabel);
        cardPanel.add(usbDirectConnectivityPanel, usbDirectBoxLabel);

        this.setLayout(new BorderLayout());
        this.add(comboBoxPanel, BorderLayout.NORTH);
        this.add(cardPanel, BorderLayout.CENTER);
    }

    private JPanel createUsbCard() {
        JPanel usbCardPanel = new JPanel();
        usbPrinterList = new JComboBox();

        getUsbPrintersAndAddToComboList();
        usbCardPanel.add(usbPrinterList);

        return usbCardPanel;
    }

    private JPanel createUsbDirectCard() {
        JPanel usbDirectCardPanel = new JPanel();
        JLabel usbDirectAddressLabel = new JLabel("USB Direct: ");
        usbDirectCardPanel.add(usbDirectAddressLabel);

        usbDirectAddressTextField = new JTextField();
        usbDirectAddressTextField.setPreferredSize(new Dimension(220, 25));
        usbDirectCardPanel.add(usbDirectAddressTextField);

        return usbDirectCardPanel;
    }

    private JPanel createNetworkCard() {
        JPanel networkPanel = new JPanel();
        JLabel ipAddressLabel = new JLabel("IP Address: ");
        networkPanel.add(ipAddressLabel);

        networkIpAddressTextField = new JTextField();
        networkIpAddressTextField.setPreferredSize(new Dimension(110, 25));
        networkPanel.add(networkIpAddressTextField);

        JLabel portNumLabel = new JLabel("Port: ");
        networkPanel.add(portNumLabel);

        networkPortNumTextField = new NumericTextField();
        networkPortNumTextField.setMaxLength(5);
        networkPortNumTextField.setPreferredSize(new Dimension(80, 25));
        networkPanel.add(networkPortNumTextField);

        return networkPanel;
    }

    private JPanel createTlsCard() {

        JPanel tlsPanel = new JPanel();
        tlsPanel.setLayout(new BoxLayout(tlsPanel, BoxLayout.Y_AXIS));

        // --------------------------------------
        // IP + Port Row (same as screenshot)
        // --------------------------------------
        JPanel ipPort = new JPanel();
        ipPort.add(new JLabel("IP Address: "));
        tlsIpAddressTextField = new JTextField();
        tlsIpAddressTextField.setPreferredSize(new Dimension(110, 25));
        ipPort.add(tlsIpAddressTextField);

        ipPort.add(new JLabel("Port: "));
        tlsPortNumTextField = new NumericTextField();
        tlsPortNumTextField.setMaxLength(5);
        tlsPortNumTextField.setPreferredSize(new Dimension(80, 25));
        ipPort.add(tlsPortNumTextField);

        tlsPanel.add(ipPort);

        // --------------------------------------
        // GROUP BOX – looks like screenshot border
        // --------------------------------------
        JPanel groupBox = new JPanel();
        groupBox.setLayout(new BoxLayout(groupBox, BoxLayout.Y_AXIS));
        groupBox.setBorder(BorderFactory.createTitledBorder("TLS Certificate Validation"));

        // FIXED SIZE
        int groupWidth = 320;
        int groupHeight = 150;

        Dimension fixedSize = new Dimension(groupWidth, groupHeight);
        groupBox.setPreferredSize(fixedSize);
        groupBox.setMinimumSize(fixedSize);
        groupBox.setMaximumSize(fixedSize);
        groupBox.setAlignmentX(Component.CENTER_ALIGNMENT);

        // Radio buttons
        trustAllRadio = new JRadioButton("None", true);
        trustJavaKsRadio = new JRadioButton("Java Keystore");
        trustCertFileRadio = new JRadioButton("CA certificate file");

        ButtonGroup group = new ButtonGroup();
        group.add(trustAllRadio);
        group.add(trustJavaKsRadio);
        group.add(trustCertFileRadio);

        groupBox.add(trustAllRadio);
        groupBox.add(trustJavaKsRadio);
        groupBox.add(trustCertFileRadio);

        // --------------------------------------
        // CERT PATH (always visible)
        // --------------------------------------
        JPanel certPathPanel = new JPanel();
        certPathPanel.setLayout(new BoxLayout(certPathPanel, BoxLayout.Y_AXIS));

        JLabel pathLabel = new JLabel("Path to CA certificate");
        tlsCertPathTextField = new JTextField();
        tlsCertPathTextField.setPreferredSize(new Dimension(150, 25));
        tlsCertBrowseButton = new JButton("Browse...");

        // Start disabled
        tlsCertPathTextField.setEnabled(false);
        tlsCertBrowseButton.setEnabled(false);

        JPanel browseRow = new JPanel();
        browseRow.add(tlsCertPathTextField);
        browseRow.add(tlsCertBrowseButton);

        certPathPanel.add(pathLabel);
        certPathPanel.add(browseRow);

        groupBox.add(certPathPanel);

        tlsPanel.add(groupBox);

        // --------------------------------------
        // Radio button behavior
        // --------------------------------------
        trustCertFileRadio.addActionListener(e -> {
            tlsCertPathTextField.setEnabled(true);
            tlsCertBrowseButton.setEnabled(true);
        });

        trustAllRadio.addActionListener(e -> {
            tlsCertPathTextField.setEnabled(false);
            tlsCertBrowseButton.setEnabled(false);
        });

        trustJavaKsRadio.addActionListener(e -> {
            tlsCertPathTextField.setEnabled(false);
            tlsCertBrowseButton.setEnabled(false);
        });

        // --------------------------------------
        // File chooser
        // --------------------------------------
        tlsCertBrowseButton.addActionListener(e -> {
            JFileChooser fc = new JFileChooser();
            fc.setFileFilter(new FileNameExtensionFilter("Certificate files", "crt", "cer", "pem"));
            if (fc.showOpenDialog(this) == JFileChooser.APPROVE_OPTION) {
                tlsCertPathTextField.setText(fc.getSelectedFile().getAbsolutePath());
            }
        });

        return tlsPanel;
    }

    public Connection getConnection() throws ConnectionException {
        int selectedIndex = comboBox.getSelectedIndex();

        switch (selectedIndex) {

        case 0: // TLS Network
            String ip = tlsIpAddressTextField.getText().trim();
            int port = getPortFromTextField(tlsPortNumTextField, 9143);

            TlsConfig tlsConfig;

            if (trustAllRadio.isSelected()) {
                tlsConfig = TlsConfig.trustAll();
            } else if (trustJavaKsRadio.isSelected()) {
                tlsConfig = TlsConfig.trustJavaKeyStore();
            } else {
                // CERT FILE
                String filePath = tlsCertPathTextField.getText().trim();
                tlsConfig = TlsConfig.fromCertificateFile(filePath);
            }

            return new TlsConnection(ip, port, tlsConfig);
        case 1: // Network (TCP)
            String networkIp = networkIpAddressTextField.getText().trim();
            int networkPort = getPortFromTextField(networkPortNumTextField, 9100);
            return new TcpConnection(networkIp, networkPort);

        case 2: // USB
            DiscoveredPrinterDriver printer = (DiscoveredPrinterDriver) usbPrinterList.getSelectedItem();
            return new DriverPrinterConnection(printer.printerName);

        case 3: // USB Direct
        default:
            return new UsbConnection(usbDirectAddressTextField.getText());
        }
    }

    private int getPortFromTextField(NumericTextField textField, int defaultPort) {
        String portText = textField.getText().trim();
        if (!portText.isEmpty()) {
            try {
                return Integer.parseInt(portText);
            } catch (NumberFormatException e) {
                // fallback to default port
            }
        }
        return defaultPort;
    }

    public void focusGained(FocusEvent event) {
        if (usbPrinterList != null) {
            int currentIndex = usbPrinterList.getSelectedIndex();
            usbPrinterList.removeAllItems();
            getUsbPrintersAndAddToComboList();
            if ((currentIndex > -1) && (currentIndex < usbPrinterList.getItemCount())) {
                usbPrinterList.setSelectedIndex(currentIndex);
            }
        }
    }

    private void getUsbPrintersAndAddToComboList() {
        DiscoveredPrinterDriver[] discoPrinters;
        try {
            discoPrinters = UsbDiscoverer.getZebraDriverPrinters();
            for (DiscoveredPrinterDriver printer : discoPrinters) {
                usbPrinterList.addItem(printer);
            }

        } catch (ConnectionException e) {
            usbPrinterList.removeAllItems();
            usbPrinterList.addItem("OS not supported");
        }
    }

    public void focusLost(FocusEvent e) {
    }

    // Optional: Method to set IP address in both network and TLS fields
    public void setIpAddress(String ipAddress) {
        networkIpAddressTextField.setText(ipAddress);
        tlsIpAddressTextField.setText(ipAddress);
    }

    // Optional: Method to get current IP address based on selection
    public String getCurrentIpAddress() {
        int selectedIndex = comboBox.getSelectedIndex();
        switch (selectedIndex) {
        case 0:
            return tlsIpAddressTextField.getText();
        case 1:
            return networkIpAddressTextField.getText();
        default:
            return "";
        }
    }
}